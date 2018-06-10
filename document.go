package bongo

import (
	"fmt"
	"reflect"
	"time"

	"github.com/globalsign/mgo/bson"
)

// Document ...
type Document interface {
	GetID() bson.ObjectId
	SetID(bson.ObjectId)
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

	// We want this to default to false without any work. So this will be the opposite of isNew. We want it to be new unless set to existing
	exists bool
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

// GetModified gets the modified date
func (d *DocumentModel) GetModified() time.Time {
	return d.Modified
}

// GetCollectionName return the value of the coll tag field if exists
func (d *DocumentModel) GetCollectionName(doc interface{}) string {
	v := reflect.ValueOf(doc).Elem()

	for i := 0; i < v.NumField(); i++ {
		fmt.Println(v.Field(i).Type().Name())
		if v.Field(i).Type().Name() == "DocumentModel" {
			fInfo := v.Type().Field(i)
			tag := fInfo.Tag
			coll := tag.Get("coll")
			return coll
		}
	}

	panic("Document model does not have a collection. Use DocumentModel and coll tag to define one")
}

// GetIndexedFields return a []string of the indexed field suitable to
// be used with EnsureIndexKey method  of mgo
func (d *DocumentModel) GetIndexedFields(doc interface{}) []string {
	v := reflect.ValueOf(doc).Elem()

	for i := 0; i < v.NumField(); i++ {
		fmt.Println(v.Field(i).Type().Name())
		if v.Field(i).Type().Name() == "DocumentModel" {
			fInfo := v.Type().Field(i)
			tag := fInfo.Tag
			idx := tag.Get("idx")
			idx = tag.Get("idx")
			fmt.Println(idx)

			return nil
		}
	}

	panic("Document model does not have a collection. Use DocumentModel and coll tag to define one")
}

func (d DocumentNotFoundError) Error() string {
	return "Document not found"
}
