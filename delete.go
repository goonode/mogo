package bongo

import (
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

// DeleteDocument ...
func (c *Collection) DeleteDocument(doc Document) error {
	var err error
	// Create a new session per mgo's suggestion to avoid blocking
	sess := c.Connection.Session.Clone()
	defer sess.Close()
	col := c.collectionOnSession(sess)

	if hook, ok := doc.(BeforeDeleteHook); ok {
		err := hook.BeforeDelete(c)
		if err != nil {
			return err
		}
	}

	err = col.Remove(bson.M{"_id": doc.GetID()})

	if err != nil {
		return err
	}

	// go CascadeDelete(c, doc)

	if hook, ok := doc.(AfterDeleteHook); ok {
		err = hook.AfterDelete(c)
		if err != nil {
			return err
		}
	}

	return nil

}

// Delete is a convenience method which just delegates to mgo. Note that hooks are NOT run
func (c *Collection) Delete(query bson.M) (*mgo.ChangeInfo, error) {
	sess := c.Connection.Session.Clone()
	defer sess.Close()
	col := c.collectionOnSession(sess)
	return col.RemoveAll(query)
}

// DeleteOne is a convenience method which just delegates to mgo. Note that hooks are NOT run
func (c *Collection) DeleteOne(query bson.M) error {
	sess := c.Connection.Session.Clone()
	defer sess.Close()
	col := c.collectionOnSession(sess)
	return col.Remove(query)
}

// DeleteDocument ...
func DeleteDocument(doc Document) error {
	return doc.GetColl().DeleteDocument(doc)
}

// Delete ... (no hooks)
func Delete(doc Document, query bson.M) (*mgo.ChangeInfo, error) {
	return doc.GetColl().Delete(query)
}

// DeleteOne ... (no hooks)
func DeleteOne(doc Document, query bson.M) error {
	return doc.GetColl().DeleteOne(query)
}
