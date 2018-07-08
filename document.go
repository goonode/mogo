package bongo

import (
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
	SetConn(c *Connection)

	// Model

	GetIName() string
	RestoreIName(n string)
}

// Document ...
type Document interface {
	Model
	GetID() bson.ObjectId
	SetID(bson.ObjectId)

	BsonID() *bson.M

	AsNew()

	SetCInfo(*mgo.ChangeInfo)
	GetCInfo() *mgo.ChangeInfo
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

	// Model index in registry
	iname string `bson:"-"`

	// Model lifecycle flags
	exists bool `bson:"-"`

	// mgo.ChangeInfo
	cinfo *mgo.ChangeInfo `bson:"-"`
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

// BsonID returns the document id using bson.M interface style
// This method can be directly used with Find, but not with FindID
// which expects directly id interface{} (i.e. d.ID/d.GetID())
func (d *DocumentModel) BsonID() *bson.M {
	return &bson.M{
		"_id": d.GetID(),
	}
}

// AsNew assign a new ID to the current document so it can
// be considered as a new document.
func (d *DocumentModel) AsNew() {
	d.ID = bson.NewObjectId()
	d.SetIsNew(true)
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

// GetColl implementation for Model interface (why may you need to change collection name ?)
func (d *DocumentModel) GetColl() *Collection {
	_, ri, ok := ModelRegistry.ExistByName(d.iname)
	if !ok {
		panic("the document model is not registered")
	}

	return DBConn.Collection(ri.Collection)
}

// GetParsedIndex return the index stored with the passed field name
func (d *DocumentModel) GetParsedIndex(name string) []ParsedIndex {
	_, ri, ok := ModelRegistry.ExistByName(d.iname)
	if !ok {
		panic("the document model is not registered")
	}

	return ri.Indexes[name]
}

// GetAllParsedIndex return all stored parsed indexes
func (d *DocumentModel) GetAllParsedIndex() map[string][]ParsedIndex {
	_, ri, ok := ModelRegistry.ExistByName(d.iname)
	if !ok {
		panic("the document model is not registered")
	}

	return ri.Indexes
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
	_, ri, ok := ModelRegistry.ExistByName(d.iname)
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

// GetIName is used to save the DocumentModel struct of the document.
// This is useful after returning from one of a find method where mgo
// driver return a freshly create zero filled struct.DocumentModel
// Note: mgo should consider bson "-" tag also on unmarshal
func (d *DocumentModel) GetIName() string {
	return d.iname
}

// RestoreIName is used to restore the DocumentModel struct
func (d *DocumentModel) RestoreIName(n string) {
	d.iname = n
}

func (d DocumentNotFoundError) Error() string {
	return "Document not found"
}

// NewDocument ...
func NewDocument(d interface{}) interface{} {
	n, ri, ok := ModelRegistry.Exists(d)
	if !ok { // Trying to register
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
	df.Set(reflect.ValueOf(dm))

	return r.Interface()
}
