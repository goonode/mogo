package mogo

import (
	"errors"
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
	GetColl() *Collection

	// Indexes

	GetParsedIndex(name string) []ParsedIndex
	GetAllParsedIndex() map[string][]ParsedIndex
	GetIndex(name string) []*mgo.Index
	GetAllIndex() []*mgo.Index

	// Refs

	GetRefIndex(name string) RefIndex

	// Connection

	GetConn() *Connection
	SetConn(*Connection)

	// Model

	GetMe() (iname string, me interface{})
	SetMe(iname string, me interface{})

	// Query

	Find(interface{}) *Query
	FindID(interface{}) *Query

	// Database Ops

	Save() error
	Remove() error
}

// Document ...
type Document interface {
	Model
	GetID() bson.ObjectId
	SetID(bson.ObjectId)

	BsonID() *bson.M

	AsNew()
	AsDocument() Document
	AsModel() Model

	SetCInfo(*mgo.ChangeInfo)
	GetCInfo() *mgo.ChangeInfo
}

// DocumentModel ...
type DocumentModel struct {
	ID       bson.ObjectId `bson:"_id,omitempty" json:"_id"`
	Created  time.Time     `bson:"_created" json:"_created"`
	Modified time.Time     `bson:"_modified" json:"_modified"`

	// Model index in registry
	iname string `bson:"-"`
	// Me
	me interface{}

	// Model lifecycle flags
	exists bool `bson:"-"`

	// mgo.ChangeInfo
	cinfo *mgo.ChangeInfo `bson:"-"`
}

// RefField is a reference field to another model. The receiver will return the real object.
type RefField struct {
	ID bson.ObjectId `bson:"_id,omitempty" json:"_id"`
}

// RefFieldSlice is a slice of RefField. The receiver will return an Iterator to this field.
type RefFieldSlice []*RefField

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

// BsonID returns the document id using bson.M interface style
// This method can be directly used with Find, but not with FindID
// which expects directly id interface{} (i.e. d.ID/d.GetID())
func (d *DocumentModel) BsonID() *bson.M {
	return &bson.M{
		"_id": d.GetID(),
	}
}

// AsNew assigns a new ID to the current document so it can
// be considered as a new document.
func (d *DocumentModel) AsNew() {
	d.ID = bson.NewObjectId()
	d.SetIsNew(true)
}

// AsDocument tranforms the DocumentModel to Document interface
func (d *DocumentModel) AsDocument() Document {
	if dI, ok := d.me.(Document); ok {
		return dI
	}

	return nil
}

// AsModel tranforms the DocumentModel to Document interface
func (d *DocumentModel) AsModel() Model {
	if dI, ok := d.me.(Model); ok {
		return dI
	}

	return nil
}

// GetCInfo gets the document cinfo field (see mgo.upsert)
func (d *DocumentModel) GetCInfo() *mgo.ChangeInfo {
	return d.cinfo
}

// SetCInfo sets the document cinfo field
func (d *DocumentModel) SetCInfo(ci *mgo.ChangeInfo) {
	d.cinfo = ci
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
	_, ri, ok := ModelRegistry.Exists(d)
	if !ok {
		panic("the document model is not registered")
	}

	return ri.Collection
}

// SetCollName implementation for Model interface (why may you need to change collection name at runtime ?)
func (d *DocumentModel) SetCollName(name string) {
}

// GetColl implementation for Model interface
func (d *DocumentModel) GetColl() *Collection {
	_, ri, ok := ModelRegistry.ExistsByName(d.iname)
	if !ok {
		panic("the document model is not registered")
	}

	return DBConn.Collection(ri.Collection)
}

// GetParsedIndex returns the index stored with the passed field name
func (d *DocumentModel) GetParsedIndex(name string) []ParsedIndex {
	_, ri, ok := ModelRegistry.ExistsByName(d.iname)
	if !ok {
		panic("the document model is not registered")
	}

	return ri.Indexes[name]
}

// GetAllParsedIndex returns all stored parsed indexes
func (d *DocumentModel) GetAllParsedIndex() map[string][]ParsedIndex {
	_, ri, ok := ModelRegistry.ExistsByName(d.iname)
	if !ok {
		panic("the document model is not registered")
	}

	return ri.Indexes
}

// GetIndex returns the mgo.Index struct required to mgo.EnsureIndex method
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

// GetAllIndex returns the mgo.Index struct required to mgo.EnsureIndex method
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

// GetRefIndex returns the RefIndex struct for the given field
func (d *DocumentModel) GetRefIndex(name string) RefIndex {
	_, ri, ok := ModelRegistry.ExistsByName(d.iname)
	if !ok {
		panic("the document model is not registered")
	}

	return ri.Refs[name]
}

// Ref same as GetRefIndex
func (d *DocumentModel) Ref(name string) RefIndex {
	_, ri, ok := ModelRegistry.ExistsByName(d.iname)
	if !ok {
		panic("the document model is not registered")
	}

	return ri.Refs[name]
}

// SetConn ...
func (d *DocumentModel) SetConn(c *Connection) {
}

// GetConn ...
func (d *DocumentModel) GetConn() *Connection {
	if ModelRegistry.Index(d.iname) == -1 {
		panic("the document model is not registered")
	}

	return DBConn
}

// GetMe is used to save the iname, me fields of the document model.
// This is useful after returning from one of a find method where mgo
// driver return a freshly create zero filled struct.DocumentModel
// Note: mgo should consider bson "-" tag also on unmarshal, actually overwrites it
func (d *DocumentModel) GetMe() (iname string, me interface{}) {
	return d.iname, d.me
}

// SetMe is used to set the iname and me fields
func (d *DocumentModel) SetMe(iname string, me interface{}) {
	d.iname = iname
	d.me = me
}

// Find is the wrapper method to mgo Find
func (d *DocumentModel) Find(query interface{}) *Query {
	q := &Query{
		MgoC:       d.GetColl().C(),
		MgoQ:       d.GetColl().C().Find(query),
		Pagination: nil,
	}

	return q
}

// FindID is a wrapper to the mgo FindId
func (d *DocumentModel) FindID(id interface{}) *Query {
	q := &Query{
		MgoC:       d.GetColl().C(),
		MgoQ:       d.GetColl().C().FindId(id),
		Pagination: nil,
	}

	return q
}

// FindOne is a shortcut for Find().One()
func (d *DocumentModel) FindOne(query interface{}, result interface{}) error {
	return d.Find(query).One(result)
}

// Populate builds a Query to populate the referenced field (see scratch)
// The returned Query object refers to the target field object not the original one.
func (d *DocumentModel) Populate(f string) *Query {
	_, i, _ := ModelRegistry.Exists(d.me)
	if i == nil { // model is not registered
		return nil
	}

	r := i.Refs[f]
	if !r.Exists { // Field name not exists
		return nil
	}

	iField := reflect.ValueOf(d.me).Elem().Field(r.Idx).Interface()
	t := ModelRegistry.New(r.Ref).(Document)
	var q = bson.M{"$populate": make([]bson.M, 0)}

	switch i.Refs[f].Kind {
	case reflect.Slice:
		var inner = bson.M{"$or": make([]bson.M, 0)}

		field := iField.(RefFieldSlice)
		for i := range field {
			inner["$or"] = append(inner["$or"].([]bson.M), bson.M{"_id": field[i].ID})
		}
		q["$populate"] = append(q["$populate"].([]bson.M), inner)
		return Find(t, q)
	default:
		field := iField.(RefField)
		var inner = bson.M{"_id": field.ID}

		q["$populate"] = append(q["$populate"].([]bson.M), inner)
		return Find(t, q)
	}

}

// FindByID is a shortcut for FindID().One()
func (d *DocumentModel) FindByID(id interface{}, result interface{}) error {
	return d.FindID(id).One(result)
}

// Save ...
func (d *DocumentModel) Save() error {
	var err error
	var cinfo *mgo.ChangeInfo

	c := d.GetColl()
	sess := c.Connection.Session.Clone()
	defer sess.Close()

	// Per mgo's recommendation, create a clone of the session so there is no blocking
	col := c.collectionOnSession(sess)

	err = c.PreSave(d.me.(Document))
	if err != nil {
		return err
	}

	// If the model implements the NewTracker interface, we'll use that to determine newness. Otherwise always assume it's new
	isNew := true
	if newt, ok := d.me.(NewTracker); ok {
		isNew = newt.IsNew()
	}

	// Add created/modified time. Also set on the model itself if it has those fields.
	now := time.Now()
	if tt, ok := d.me.(TimeCreatedTracker); ok && isNew {
		tt.SetCreated(now)
	}

	if tt, ok := d.me.(TimeModifiedTracker); ok {
		tt.SetModified(now)
	}

	// If the model has indexes we create them here...
	idxs := d.GetAllIndex()
	for i := range idxs {
		err = col.EnsureIndex(*idxs[i])
		if err != nil {
			return err
		}
	}

	id := d.GetID()
	if !isNew && !id.Valid() {
		return errors.New("New tracker says this document isn't new but there is no valid Id field")
	}

	if isNew && !id.Valid() {
		// Generate an Id
		id = bson.NewObjectId()
		d.SetID(id)
	}

	cinfo, err = col.UpsertId(id, d.me)
	d.SetCInfo(cinfo)

	if err != nil {
		return err
	}

	if hook, ok := d.me.(AfterSaveHook); ok {
		err = hook.AfterSave()
		if err != nil {
			return err
		}
	}

	// We saved it, no longer new
	if newt, ok := d.me.(NewTracker); ok {
		newt.SetIsNew(false)
	}

	return nil
}

// Remove removes document from database, running
// before and after delete hooks
func (d *DocumentModel) Remove() error {
	var err error
	// Create a new session per mgo's suggestion to avoid blocking
	c := d.GetColl()
	sess := c.Connection.Session.Clone()
	defer sess.Close()
	col := c.collectionOnSession(sess)

	if hook, ok := d.me.(BeforeDeleteHook); ok {
		err := hook.BeforeDelete()
		if err != nil {
			return err
		}
	}

	err = col.RemoveId(d.GetID())

	if err != nil {
		return err
	}

	if hook, ok := d.me.(AfterDeleteHook); ok {
		err = hook.AfterDelete()
		if err != nil {
			return err
		}
	}

	return nil
}

// NewDoc ...
func NewDoc(d interface{}) interface{} {
	var n string
	var ri *ModelInternals
	var ok bool

	n, ri, ok = ModelRegistry.ExistsByName(interfaceName(d))
	if !ok { // Trying to register (?to be removed?)
		ModelRegistry.Register(d)
		n, ri, _ = ModelRegistry.Exists(d)
	}
	t := ri.Type
	v := ValueOf(d)
	i := ModelRegistry.Index(n) // The dm

	r := reflect.New(t)
	r.Elem().Set(v)

	df := r.Elem().Field(i)
	dm := df.Interface().(DocumentModel)
	dm.iname = n
	dm.me = r.Interface()
	df.Set(reflect.ValueOf(dm))

	return r.Interface()
}

// MakeDoc (no returning new) adds the DocumentModel to an existant interface ((TODO))
func MakeDoc(d interface{}) interface{} {
	return nil
}
