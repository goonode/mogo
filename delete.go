package mogo

import (
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

// Remove removes the passed document from database, executing
// before and after delete hooks
func (c *Collection) Remove(doc Document) error {
	var err error
	// Create a new session per mgo's suggestion to avoid blocking
	sess := c.Connection.Session.Clone()
	defer sess.Close()
	col := c.collectionOnSession(sess)

	if hook, ok := doc.(BeforeDeleteHook); ok {
		err := hook.BeforeDelete()
		if err != nil {
			return err
		}
	}

	err = col.RemoveId(doc.GetID())

	if err != nil {
		return err
	}

	if hook, ok := doc.(AfterDeleteHook); ok {
		err = hook.AfterDelete()
		if err != nil {
			return err
		}
	}

	return nil
}

// RemoveAll removes all documents passed in slice executing,
// for each one, the before and after delete hooks.
// Document in slice can belong to different collections.
func (c *Collection) RemoveAll(docs []Document) map[bson.ObjectId]error {
	var err error
	var errs = make(map[bson.ObjectId]error, 0)
	var col *mgo.Collection

	// Create a new session per mgo's suggestion to avoid blocking
	sess := c.Connection.Session.Clone()
	defer sess.Close()

	for _, d := range docs {
		col = d.GetColl().collectionOnSession(sess)

		if hook, ok := d.(BeforeDeleteHook); ok {
			err = hook.BeforeDelete()
			if err != nil {
				errs[d.GetID()] = err
				continue
			}
		}
		err = col.Remove(d.BsonID)
		if err != nil {
			errs[d.GetID()] = err
			continue
		}

		if hook, ok := d.(AfterDeleteHook); ok {
			err = hook.AfterDelete()
			if err != nil {
				errs[d.GetID()] = err
				continue
			}
		}
	}

	if l := len(errs); l == 0 {
		return nil
	}

	return errs
}

// RemoveBySelector is a wrapper aorund mgo.Remove method.
// This is faster the Collection.Remove method but it doesn't run
// before / after delete hooks
func (c *Collection) RemoveBySelector(selector interface{}) error {
	var err error
	// Create a new session per mgo's suggestion to avoid blocking
	sess := c.Connection.Session.Clone()
	defer sess.Close()
	col := c.collectionOnSession(sess)

	err = col.Remove(selector)

	if err != nil {
		return err
	}

	return nil
}

// RemoveAllBySelector is a wrapper aorund mgo.RemoveAll method.
//
// This is faster then Collection.RemoveAll method but it doesn't
// run before / after delete hooks.
//
// The selectors argument is a map[Model]interface{},
// where the key is the model (collection) on which we are going
// to apply the selector and the interface is the mgo selector.
//
//
func (c *Collection) RemoveAllBySelector(selectors map[Model]interface{}) map[string]*ChangeInfoWithError {
	var err error
	var info *mgo.ChangeInfo
	var errs = make(map[string]*ChangeInfoWithError)

	// Create a new session per mgo's suggestion to avoid blocking
	sess := c.Connection.Session.Clone()
	defer sess.Close()

	for m, s := range selectors {
		col := m.GetColl().collectionOnSession(sess)
		info, err = col.RemoveAll(s)

		if err != nil {
			iname, _ := m.GetMe()
			errs[iname] = &ChangeInfoWithError{Info: info, Err: err}
		}
	}

	if len(errs) > 0 {
		return errs
	}

	return nil
}

// Remove is a convenience (haha) method for Document.Remove
func Remove(doc Document) error {
	return doc.Remove()
}

// RemoveAll is convenience method for Collection.RemoveAll
func RemoveAll(docs []Document) map[bson.ObjectId]error {
	if l := len(docs); l > 0 {
		d := docs[0]
		return d.GetColl().RemoveAll(docs)
	}

	return nil
}

// RemoveBySelector is convenience method for Collection.RemoveBySelector
// The model argument here is used to reference its Collection to call
// the mgo driver on it.
func RemoveBySelector(model Model, selector interface{}) error {
	return model.GetColl().RemoveBySelector(selector)
}

// RemoveAllBySelector is convenience method for Collection.RemoveAllBySelector
func RemoveAllBySelector(selectors map[Model]interface{}) map[string]*ChangeInfoWithError {
	for m := range selectors {
		return m.GetColl().RemoveAllBySelector(selectors)
	}

	return nil
}
