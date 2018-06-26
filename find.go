package bongo

import (
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

// FindOne ...
func (c *Collection) FindOne(query interface{}, doc Model) error {

	// Now run a find
	results := c.Find(query)
	results.Query.Limit(1)

	hasNext := results.Next(doc)

	if !hasNext {
		// There could have been an error fetching the next one, which would set the Error property on the resultset
		if results.Error != nil {
			return results.Error
		}
		return &DocumentNotFoundError{}
	}

	if newt, ok := doc.(NewTracker); ok {
		newt.SetIsNew(false)
	}

	return nil
}

// FindByID ...
func (c *Collection) FindByID(id bson.ObjectId, doc Model) error {
	iname := doc.SaveIName()
	err := c.Collection().FindId(id).One(doc)
	doc.RestoreIName(iname)
	// Handle errors coming from mgo - we want to convert it to a DocumentNotFoundError so people can figure out
	// what the error type is without looking at the text
	if err != nil {
		if err == mgo.ErrNotFound {
			return &DocumentNotFoundError{}
		}
		return err
	}

	if hook, ok := doc.(AfterFindHook); ok {
		err = hook.AfterFind(c)
		if err != nil {
			return err
		}
	}

	// We retrieved it, so set new to false
	if newt, ok := doc.(NewTracker); ok {
		newt.SetIsNew(false)
	}
	return nil
}

// Find doesn't actually do any DB interaction, it just creates the result set so we can
// start looping through on the iterator
func (c *Collection) Find(query interface{}) *ResultSet {
	col := c.Collection()

	// Count for testing
	q := col.Find(query)

	resultset := new(ResultSet)

	resultset.Query = q
	resultset.Params = query
	resultset.Collection = c

	return resultset
}

// FindOne ...
func FindOne(doc Document, query interface{}) error {
	return doc.GetColl().FindOne(query, doc)
}

// FindByID ...
func FindByID(doc Document, id bson.ObjectId) error {
	return doc.GetColl().FindByID(id, doc)
}

// Find ...
func Find(doc Model, query interface{}) *ResultSet {
	if d, ok := doc.(Document); ok {
		return d.GetColl().Find(query)
	}

	return nil
}
