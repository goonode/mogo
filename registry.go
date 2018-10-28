package mogo

import (
	"errors"
	"fmt"
	"log"
	"reflect"
	"sync"

	"github.com/globalsign/mgo"
)

// Config ...
type Config struct {
	ConnectionString string
	Database         string
	DialInfo         *mgo.DialInfo
}

// var EncryptionKey [32]byte
// var EnableEncryption bool

// Connection ...
type Connection struct {
	Config  *Config
	Session *mgo.Session
	Context *Context
}

// Registry ...
type Registry interface {
	Register(...interface{})
	Exists(interface{}) (string, *ModelInternals, bool)
	ExistByName(string) (string, *ModelInternals, bool)

	Index(string) int
	TypeOf(string) reflect.Type

	New(string) interface{}

	Field(string, interface{}) interface{}
}

// ModelInternals contains some internal information about the model
type ModelInternals struct {
	// Idx is the index of the field containing the DM
	Idx int
	// The Type
	Type reflect.Type

	// Model internal data
	Collection string
	Indexes    map[string][]ParsedIndex
	Refs       map[string]RefIndex
}

// ModelReg ...
type ModelReg map[string]*ModelInternals

// ModelRegistry is the centralized registry of all models used for the app
var ModelRegistry = make(ModelReg, 0)

// DBConn is the connection initialized after Connect is called.
// All underlying operations are made using this connection
var DBConn *Connection

var mu sync.Mutex

// Connect creates a new connection and run Connect()
func Connect(config *Config) (*Connection, error) {
	conn := &Connection{
		Config:  config,
		Context: &Context{},
	}

	err := conn.Connect()

	if err != nil {
		DBConn = nil
		log.Printf("Error while connectiong to MongoDb (err: %v)", err)

		return nil, err
	}

	DBConn = conn
	return conn, err
}

// Register ...
func (r ModelReg) Register(i ...interface{}) {
	defer mu.Unlock()

	mu.Lock()
	for p, o := range i {
		t := reflect.TypeOf(o)
		v := reflect.ValueOf(o)

		if t.Kind() == reflect.Ptr {
			t = reflect.Indirect(reflect.ValueOf(o)).Type()
			v = reflect.ValueOf(o).Elem()
		}
		n := t.Name()
		if t.Kind() != reflect.Struct {
			panic(fmt.Sprintf("Only type struct can be used as document model (passed type %s (pos: %d) is not struct)", n, p))
		}
		var idx = -1
		for i := 0; i < v.NumField(); i++ {
			ft := t.Field(i)
			if ft.Type.ConvertibleTo(reflect.TypeOf(DocumentModel{})) {
				idx = i
				break
			}
		}

		if idx == -1 {
			panic(fmt.Sprintf("A document model must embed a DocumentModel type field (passed type %s (pos: %d) does not have)", n, p))
		}

		pi, refs, coll := initializeTags(t, v)
		if coll == "" {
			panic(fmt.Sprintf("The document model does not have a collection name (passed type %s)", n))
		}

		ModelRegistry[n] = &ModelInternals{
			Idx:        idx,
			Type:       t,
			Collection: coll,
			Indexes:    pi,
			Refs:       refs}
	}

	for k, v := range ModelRegistry {
		// TODO: Second Pass to validate all defined Refs
		for kk, vv := range v.Refs {
			if !vv.Exists {
				if _, ok := ModelRegistry[vv.Ref]; ok {
					ModelRegistry[k].Refs[kk] = RefIndex{
						Model:  k,
						Idx:    ModelRegistry[k].Refs[kk].Idx,
						Ref:    ModelRegistry[k].Refs[kk].Ref,
						Kind:   ModelRegistry[k].Refs[kk].Kind,
						Type:   ModelRegistry[k].Refs[kk].Type,
						Exists: true,
					}
				}
			}
		}
	}
}

// Exists ...
func (r ModelReg) Exists(i interface{}) (string, *ModelInternals, bool) {
	t := reflect.TypeOf(i)
	if t.Kind() == reflect.Ptr {
		t = reflect.Indirect(reflect.ValueOf(i)).Type()
	}
	n := t.Name()

	if rT, ok := ModelRegistry[n]; ok {
		return n, rT, true
	}
	return "", nil, false
}

// ExistsByName ...
func (r ModelReg) ExistsByName(n string) (string, *ModelInternals, bool) {
	if t, ok := ModelRegistry[n]; ok {
		return n, t, true
	}
	return "", nil, false
}

// TypeOf ...
func (r ModelReg) TypeOf(n string) reflect.Type {
	if v, ok := ModelRegistry[n]; ok {
		return v.Type
	}
	return nil
}

// Index returns the index of the DocumentModel field in the struct
// or -1 if the struct name passed is not found
func (r ModelReg) Index(n string) int {
	if v, ok := ModelRegistry[n]; ok {
		return v.Idx
	}
	return -1
}

// Refs returns the Refs of the DocumentModel field in the struct
// or nil if the struct name passed is not found
func (r ModelReg) Refs(n string) map[string]RefIndex {
	if v, ok := ModelRegistry[n]; ok {
		return v.Refs
	}

	return nil
}

// SearchRef performs a search for n in Refs map and returns the *ModelInternals
// and *RefIndex if found it, or nil if not found.
func (r ModelReg) SearchRef(i interface{}, n string) (*ModelInternals, *RefIndex) {
	if _, v, ok := r.Exists(i); ok {
		for k, vv := range v.Refs {
			if k == n {
				return v, &vv
			}
		}
	}

	return nil, nil
}

// New ...
func (r ModelReg) New(n string) interface{} {
	if n, m, ok := ModelRegistry.ExistsByName(n); ok {
		v := reflect.New(m.Type)

		df := v.Elem().Field(m.Idx)
		d := df.Interface().(DocumentModel)
		d.iname = n
		df.Set(reflect.ValueOf(d))

		return v.Interface()
	}

	return nil
}

// Field meturn the field the passed document model
func (r ModelReg) Field(i int, d interface{}) reflect.Value {
	return reflect.ValueOf(d).Elem().Field(i)
}

// Connect to the database using the provided config
func (m *Connection) Connect() (err error) {
	defer func() {
		if r := recover(); r != nil {
			// panic(r)
			// return
			if e, ok := r.(error); ok {
				err = e
			} else if e, ok := r.(string); ok {
				err = errors.New(e)
			} else {
				err = errors.New(fmt.Sprint(r))
			}

		}
	}()

	if m.Config.DialInfo == nil {
		if m.Config.DialInfo, err = mgo.ParseURL(m.Config.ConnectionString); err != nil {
			panic(fmt.Sprintf("cannot parse given URI %s due to error: %s", m.Config.ConnectionString, err.Error()))
		}
	}

	session, err := mgo.DialWithInfo(m.Config.DialInfo)
	if err != nil {
		return err
	}

	m.Session = session

	m.Session.SetMode(mgo.Monotonic, true)

	return nil
}

// CollectionFromDatabase ...
func (m *Connection) CollectionFromDatabase(name string, database string) *Collection {
	// Just create a new instance - it's cheap and only has name and a database name
	return &Collection{
		Connection: m,
		Context:    m.Context,
		Database:   database,
		Name:       name,
	}
}

// Collection ...
func (m *Connection) Collection(name string) *Collection {
	return m.CollectionFromDatabase(name, m.Config.Database)
}

func buildRefIndex(idx int, tag string, fname string, t reflect.Type) RefIndex {
	if tag != "" {
		if ModelRegistry.Index(tag) == -1 {
			return RefIndex{
				Idx:    idx,
				Ref:    tag,
				Kind:   t.Kind(),
				Type:   t,
				Exists: false,
			}
		}

		return RefIndex{
			Idx:    idx,
			Ref:    tag,
			Kind:   t.Kind(),
			Type:   t,
			Exists: true,
		}
	}

	panic(fmt.Sprintf("ref tag is missing on RefField field (type: %s)", fname))
}

func initializeTags(t reflect.Type, v reflect.Value) (map[string][]ParsedIndex, map[string]RefIndex, string) {
	var coll = ""
	var pi = make(map[string][]ParsedIndex, 0)
	var ref = make(map[string]RefIndex, 0)

	for i := 0; i < v.NumField(); i++ {
		// f := v.Field(i)
		ft := t.Field(i)
		// n := "_" + ft.Name
		switch ft.Type.Kind() {
		case reflect.Struct:
			if ft.Type.ConvertibleTo(reflect.TypeOf(DocumentModel{})) {
				coll = extractColl(ft)
				pi[ft.Type.Name()] = IndexScan(extractIdx(ft))
				break
			}
			if ft.Type.ConvertibleTo(reflect.TypeOf(RefField{})) {
				r := buildRefIndex(i, extractRef(ft), ft.Name, ft.Type)
				ref[ft.Name] = r
			}
			fallthrough
		case reflect.Slice:
			if ft.Type.ConvertibleTo(reflect.TypeOf([]RefField{})) || ft.Type.ConvertibleTo(reflect.TypeOf([]*RefField{})) {
				r := buildRefIndex(i, extractRef(ft), t.Name(), ft.Type)
				ref[ft.Name] = r
			}
			fallthrough
		default:
			pi[ft.Name] = IndexScan(extractIdx(ft))
			logBadColl(ft)
		}
	}

	return pi, ref, coll
}

func logBadColl(sf reflect.StructField) {
	if extractColl(sf) != "" {
		log.Printf("Tag collection used outside DocumentModel is ignored (field: %s)", sf.Name)
	}
}

func extractColl(sf reflect.StructField) string {
	coll := sf.Tag.Get("coll")
	if coll == "" {
		coll = sf.Tag.Get("collection")
	}

	return coll
}

func extractIdx(sf reflect.StructField) string {
	idx := sf.Tag.Get("idx")
	if idx == "" {
		idx = sf.Tag.Get("index")
	}

	return idx
}

func extractRef(sf reflect.StructField) string {
	ref := sf.Tag.Get("ref")
	if ref == "" {
		ref = sf.Tag.Get("reference")
	}

	return ref
}

func interfaceName(i interface{}) string {
	var n string

	v := reflect.ValueOf(i)
	switch v.Kind() {
	case reflect.Slice, reflect.Map:
		inner := v.Type().Elem()
		switch inner.Kind() {
		case reflect.Ptr:
			n = v.Type().Elem().Elem().Name()
		default:
			n = v.Type().Elem().Name()
		}
	case reflect.Ptr:
		return interfaceName(reflect.Indirect(v).Interface())
	default:
		n = v.Type().Name()
	}

	return n
}
