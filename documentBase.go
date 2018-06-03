package bongo

import (
	"time"

	"github.com/globalsign/mgo/bson"
)

type DocumentBase struct {
	Id       bson.ObjectId `bson:"_id,omitempty" json:"_id"`
	Created  time.Time     `bson:"_created" json:"_created"`
	Modified time.Time     `bson:"_modified" json:"_modified"`

	// We want this to default to false without any work. So this will be the opposite of isNew. We want it to be new unless set to existing
	exists bool
}

// SetIsNew satisfies the new tracker interface
func (d *DocumentBase) SetIsNew(isNew bool) {
	d.exists = !isNew
}

// IsNew to ask Is the document new
func (d *DocumentBase) IsNew() bool {
	return !d.exists
}

// GetId satisfies the document interface
func (d *DocumentBase) GetId() bson.ObjectId {
	return d.Id
}

// SetId sets the ID for the document
func (d *DocumentBase) SetId(id bson.ObjectId) {
	d.Id = id
}

// SetCreated sets the created date
func (d *DocumentBase) SetCreated(t time.Time) {
	d.Created = t
}

// GetCreated gets the created date
func (d *DocumentBase) GetCreated() time.Time {
	return d.Created
}

// SetModified sets the modified date
func (d *DocumentBase) SetModified(t time.Time) {
	d.Modified = t
}

// GetModified gets the modified date
func (d *DocumentBase) GetModified() time.Time {
	return d.Modified
}
