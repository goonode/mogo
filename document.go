package bongo

import (
	"errors"
	"fmt"
	"log"
	"reflect"
	"time"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

// Model ...
type Model interface {
	// Collection

	GetCollName() string
	SetCollName(name string)
	GetCollection() *Collection

	// Indexes

	GetParsedIndex(name string) []ParsedIndex
	GetAllParsedIndex() map[string][]ParsedIndex
	GetIndex(name string) []*mgo.Index
	GetAllIndex() []*mgo.Index

	// Refs
	GetRefIndex(name string) RefIndex

	// Connection

	GetConnection() *Connection
	SetConnection(c *Connection)

	// Model

	SaveModel() DocumentModel
	RestoreModel(d DocumentModel)
}

// Document ...
type Document interface {
	Model
	GetID() bson.ObjectId
	SetID(bson.ObjectId)

	MakeAsNew()
}

// RefField ...
type RefField struct {
	ID bson.ObjectId `bson:"_id,omitempty" json:"_id"`
}

// CascadingDocument ...
type CascadingDocument interface {
	GetCascade(*Collection) []*CascadeConfig
}

// DocumentNotFoundError ...
type DocumentNotFoundError struct{}

// DocumentModel ...
type DocumentModel struct {
	ID       bson.ObjectId `bson:"_id,omitempty" json:"_id"`
	Created  time.Time     `bson:"_created" json:"_created"`
	Modified time.Time     `bson:"_modified" json:"_modified"`

	// Model internal data
	collection string                   `bson:"-"`
	index      map[string][]ParsedIndex `bson:"-"`
	refs       map[string]RefIndex      `bson:"-"`
	modelType  reflect.Type             `bson:"-"`

	// Connection
	connection *Connection `bson:"-"`

	// Model lifecycle flags
	exists      bool `bson:"-"`
	initialized bool `bson:"-"`
}

// SetIsNew satisfies the new tracker interface
func (d *DocumentModel) SetIsNew(isNew bool) {
	d.exists = !isNew
}

// IsNew to ask Is the document new
func (d *DocumentModel) IsNew() bool {
	return !d.exists
}

// GetID satisfies the document interface
func (d *DocumentModel) GetID() bson.ObjectId {
	return d.ID
}

// SetID sets the ID for the document
func (d *DocumentModel) SetID(id bson.ObjectId) {
	d.ID = id
}

// MakeAsNew assign a new ID to the current document so it can
// be considered as a new document.
func (d *DocumentModel) MakeAsNew() {
	d.ID = bson.NewObjectId()
	d.SetIsNew(true)
}

// SetCreated sets the created date
func (d *DocumentModel) SetCreated(t time.Time) {
	d.Created = t
}

// GetCreated gets the created date
func (d *DocumentModel) GetCreated() time.Time {
	return d.Created
}

// SetModified sets the modified date
func (d *DocumentModel) SetModified(t time.Time) {
	d.Modified = t
}

// The Model interface implementation

// GetModified gets the modified date
func (d *DocumentModel) GetModified() time.Time {
	return d.Modified
}

// GetCollName implementation for the Model interface
func (d *DocumentModel) GetCollName() string {
	if !d.initialized {
		panic("This document model is not initialized. Call NewDocumentModel on type first")
	}

	return d.collection
}

// SetCollName implementation for Model interface (why may you need to change collection name ?)
func (d *DocumentModel) SetCollName(name string) {
	if !d.initialized {
		panic("This document model is not initialized. Call NewDocumentModel on type first")
	}

	d.collection = name
}

// GetCollection implementation for Model interface (why may you need to change collection name ?)
func (d *DocumentModel) GetCollection() *Collection {
	if !d.initialized {
		panic("This document model is not initialized. Call NewDocumentModel on type first")
	}

	return d.connection.Collection(d.collection)
}

// GetParsedIndex return the index stored with the passed field name
func (d *DocumentModel) GetParsedIndex(name string) []ParsedIndex {
	if !d.initialized {
		panic("This document model is not initialized. Call NewDocumentModel on type first")
	}

	return d.index[name]
}

// GetAllParsedIndex return all stored parsed indexes
func (d *DocumentModel) GetAllParsedIndex() map[string][]ParsedIndex {
	if !d.initialized {
		panic("This document model is not initialized. Call NewDocumentModel on type first")
	}

	return d.index
}

// GetIndex return the mgo.Index struct required to mgo.EnsureIndex method
// using the ParsedIndex information stored for passed field name.
// TODO: discard bad formatted indexes
func (d *DocumentModel) GetIndex(name string) []*mgo.Index {
	mi := []*mgo.Index{}

	if pi := d.GetParsedIndex(name); pi != nil {
		for i := range pi {
			mi = append(mi, BuildIndex(pi[i]))
		}
		return mi
	}

	return nil
}

// GetAllIndex return the mgo.Index struct required to mgo.EnsureIndex method
// using the ParsedIndex information stored in the index map of the Model.
// TODO: discard bad formatted indexes
func (d *DocumentModel) GetAllIndex() []*mgo.Index {
	mi := []*mgo.Index{}

	if mpi := d.GetAllParsedIndex(); mpi != nil {
		for _, v := range mpi {
			for i := range v {
				mi = append(mi, BuildIndex(v[i]))
			}
		}
		return mi
	}

	return nil
}

// GetRefIndex return the RefIndex struct for the given field
func (d *DocumentModel) GetRefIndex(name string) RefIndex {
	if !d.initialized {
		panic("This document model is not initialized. Call NewDocument on type first")
	}

	return d.refs[name]
}

// SetConnection ...
func (d *DocumentModel) SetConnection(c *Connection) {
	if !d.initialized {
		panic("This document model is not initialized. Call NewDocument on type first")
	}

	d.connection = c
}

// GetConnection ...
func (d *DocumentModel) GetConnection() *Connection {
	if !d.initialized {
		panic("This document model is not initialized. Call NewDocument on type first")
	}

	return d.connection
}

// SaveModel is used to save the DocumentModel struct of the document.
// This is useful after returning from one of a find method where mgo
// driver return a freshly create zero filled struct.DocumentModel
// Note: mgo should consider bson "-" tag also on unmarshal
func (d *DocumentModel) SaveModel() DocumentModel {
	if !d.initialized {
		panic("This document model is not initialized. Call NewDocument on type first")
	}

	newDocumentModel := DocumentModel{
		collection:  d.collection,
		initialized: d.initialized,
		exists:      d.exists,
		modelType:   d.modelType,
		index:       map[string][]ParsedIndex{},
		connection:  d.connection,
	}

	for k, v := range d.index {
		newDocumentModel.index[k] = v
	}

	return newDocumentModel
}

// RestoreModel is used to restore the DocumentModel struct
func (d *DocumentModel) RestoreModel(o DocumentModel) {
	d.collection = o.collection
	d.initialized = o.initialized
	d.exists = o.exists
	d.modelType = o.modelType
	d.connection = o.connection

	d.index = map[string][]ParsedIndex{}
	for k, v := range o.index {
		d.index[k] = v
	}
}

// Helpers functions

// Save ...
func Save(doc interface{}) error {
	if d, ok := doc.(Document); ok {
		return d.GetCollection().Save(d)
	}

	return errors.New("passed document does not implement Document interface")
}

// Find ...
func Find(doc Model, query interface{}) *ResultSet {
	if d, ok := doc.(Document); ok {
		return d.GetCollection().Find(query)
	}

	return nil
}

// FindByID ...
func FindByID(doc Document, id bson.ObjectId) error {
	return doc.GetCollection().FindByID(id, doc)
}

// FindOne ...
func FindOne(doc Document, query interface{}) error {
	return doc.GetCollection().FindOne(query, doc)
}

// DeleteDocument ...
func DeleteDocument(doc Document) error {
	return doc.GetCollection().DeleteDocument(doc)
}

func (d DocumentNotFoundError) Error() string {
	return "Document not found"
}

func initializeModel(t reflect.Type, v reflect.Value) (map[string][]ParsedIndex, string) {
	var coll = ""
	var pi = make(map[string][]ParsedIndex, 0)

	for i := 0; i < v.NumField(); i++ {
		f := v.Field(i)
		ft := t.Field(i)
		n := "_" + ft.Name
		switch ft.Type.Kind() {
		case reflect.Map:
			pi[n] = IndexScan(ft.Tag.Get("idx"))
			if f.CanSet() {
				f.Set(reflect.MakeMap(ft.Type))
			}
		case reflect.Slice:
			pi[n] = IndexScan(ft.Tag.Get("idx"))
			if f.CanSet() {
				f.Set(reflect.MakeSlice(ft.Type, 0, 0))
			}
		case reflect.Chan:
			pi[n] = IndexScan(ft.Tag.Get("idx"))
			if f.CanSet() {
				f.Set(reflect.MakeChan(ft.Type, 0))
			}
		case reflect.Struct:
			if ft.Type.ConvertibleTo(reflect.TypeOf(DocumentModel{})) {
				coll = ft.Tag.Get("coll")
				pi[ft.Type.Name()] = IndexScan(ft.Tag.Get("idx"))
				break
			}
			rpi, _ := initializeModel(ft.Type, f)
			for k, v := range rpi {
				nn := n + k
				pi[nn] = v
			}
		case reflect.Ptr:
			if f.CanSet() {
				fv := reflect.New(ft.Type.Elem())
				rpi, _ := initializeModel(ft.Type.Elem(), fv.Elem())
				for k, v := range rpi {
					nn := n + k
					pi[nn] = v
				}
				f.Set(fv)
			}
		default:
			pi[n] = IndexScan(ft.Tag.Get("idx"))
		}
	}

	return pi, coll
}

func getRefIndex(idx int, tag string, fname string) RefIndex {
	if tag != "" {
		if modelRegistry.Index(tag) == -1 {
			panic(fmt.Sprintf("passed ref (%s) does not exist in registry", tag))
		}
		return RefIndex{
			Idx: idx,
			Ref: tag,
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
				coll = ft.Tag.Get("coll")
				pi[ft.Type.Name()] = IndexScan(ft.Tag.Get("idx"))
				break
			}
			if ft.Type.ConvertibleTo(reflect.TypeOf(RefField{})) {
				r := getRefIndex(i, ft.Tag.Get("ref"), ft.Name)
				ref[ft.Name] = r
			}
			fallthrough
		case reflect.Slice:
			if ft.Type.ConvertibleTo(reflect.TypeOf([]RefField{})) {
				r := getRefIndex(i, ft.Tag.Get("ref"), t.Name())
				ref[ft.Name] = r
			}
			fallthrough
		default:
			pi[ft.Name] = IndexScan(ft.Tag.Get("idx"))
			logBadColl(ft)
		}
	}

	return pi, ref, coll
}

// NewDocument ...
func NewDocument(d interface{}, c *Connection) interface{} {
	n, ri, ok := modelRegistry.Exists(d)
	if !ok { // Trying to register
		modelRegistry.Register(d)
		n, ri, _ = modelRegistry.Exists(d)
	}
	t := ri.Type
	v := reflect.ValueOf(d)
	i := modelRegistry.Index(n) // The dm

	dv := v.Field(i)
	doc := dv.Interface().(DocumentModel)

	// This document is already initialized so just creating new object
	// and assigning DM to it
	if doc.initialized {
		r := reflect.New(t)
		df := r.Elem().Field(i)
		df.Set(dv)

		return r.Interface()
	}

	var coll string
	var pi map[string][]ParsedIndex
	var refs map[string]RefIndex

	r := reflect.New(t)
	pi, refs, coll = initializeTags(t, v)
	if coll == "" {
		panic(fmt.Sprintf("The document model does not have a collection name (passed type %s)", n))
	}

	r.Elem().Set(v)
	df := r.Elem().Field(i)
	dm := df.Interface().(DocumentModel)

	dm.modelType = t
	dm.initialized = true
	dm.collection = coll
	dm.index = pi
	dm.refs = refs
	dm.connection = c

	df.Set(reflect.ValueOf(dm))

	return r.Interface()
}

func logBadColl(sf reflect.StructField) {
	if sf.Tag.Get("coll") != "" {
		log.Printf("Tag 'coll' used outside DocumentModel is ignored (field: %s)", sf.Name)
	}
}
