package mogo

import (
	"errors"
	"time"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

// PreSave ...
func (c *Collection) PreSave(doc Document) error {
	// Validate?
	if validator, ok := doc.(ValidateHook); ok {
		errs := validator.Validate()

		if len(errs) > 0 {
			return &ValidationError{errs}
		}
	}

	if hook, ok := doc.(BeforeSaveHook); ok {
		err := hook.BeforeSave()
		if err != nil {
			return err
		}
	}

	return nil
}

// Save ...
func (c *Collection) Save(doc Document) error {
	var err error
	var cinfo *mgo.ChangeInfo

	sess := c.Connection.Session.Clone()
	defer sess.Close()

	// Per mgo's recommendation, create a clone of the session so there is no blocking
	col := c.collectionOnSession(sess)

	err = c.PreSave(doc)
	if err != nil {
		return err
	}

	// If the model implements the NewTracker interface, we'll use that to determine newness. Otherwise always assume it's new
	isNew := true
	if newt, ok := doc.(NewTracker); ok {
		isNew = newt.IsNew()
	}

	// Add created/modified time. Also set on the model itself if it has those fields.
	now := time.Now()
	if tt, ok := doc.(TimeCreatedTracker); ok && isNew {
		tt.SetCreated(now)
	}

	if tt, ok := doc.(TimeModifiedTracker); ok {
		tt.SetModified(now)
	}

	// If the model has indexes we create them here...
	idxs := doc.GetAllIndex()
	for i := range idxs {
		err = col.EnsureIndex(*idxs[i])
		if err != nil {
			return err
		}
	}

	// go CascadeSave(c, doc)

	id := doc.GetID()

	if !isNew && !id.Valid() {
		return errors.New("New tracker says this document isn't new but there is no valid Id field")
	}

	if isNew && !id.Valid() {
		// Generate an Id
		id = bson.NewObjectId()
		doc.SetID(id)
	}

	cinfo, err = col.UpsertId(id, doc)
	doc.SetCInfo(cinfo)

	if err != nil {
		return err
	}

	if hook, ok := doc.(AfterSaveHook); ok {
		err = hook.AfterSave()
		if err != nil {
			return err
		}
	}

	// We saved it, no longer new
	if newt, ok := doc.(NewTracker); ok {
		newt.SetIsNew(false)
	}

	return nil
}

// Save helper function
func Save(doc Document) error {
	return doc.GetColl().Save(doc)
}
