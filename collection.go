package mogo

import (
	"strings"
	"time"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

// ChangeInfoWithError is a return value for most of mgo methods
type ChangeInfoWithError struct {
	Info *mgo.ChangeInfo
	Err  error
}

// Collection ...
type Collection struct {
	Name       string
	Database   string
	Context    *Context
	Connection *Connection
}

// BeforeSaveHook ...
type BeforeSaveHook interface {
	BeforeSave() error
}

// AfterSaveHook ...
type AfterSaveHook interface {
	AfterSave() error
}

// BeforeDeleteHook ...
type BeforeDeleteHook interface {
	BeforeDelete() error
}

// AfterDeleteHook ...
type AfterDeleteHook interface {
	AfterDelete() error
}

// AfterFindHook ...
type AfterFindHook interface {
	AfterFind() error
}

// ValidateHook ...
type ValidateHook interface {
	Validate() []error
}

// ValidationError ...
type ValidationError struct {
	Errors []error
}

// TimeCreatedTracker ...
type TimeCreatedTracker interface {
	GetCreated() time.Time
	SetCreated(time.Time)
}

// TimeModifiedTracker ...
type TimeModifiedTracker interface {
	GetModified() time.Time
	SetModified(time.Time)
}

// NewTracker ...
type NewTracker interface {
	SetIsNew(bool)
	IsNew() bool
}

func (v *ValidationError) Error() string {
	errs := make([]string, len(v.Errors))

	for i, e := range v.Errors {
		errs[i] = e.Error()
	}
	return "Validation failed. (" + strings.Join(errs, ", ") + ")"
}

// C ...
func (c *Collection) C() *mgo.Collection {
	return c.Connection.Session.DB(c.Database).C(c.Name)
}

// collectionOnSession ...
func (c *Collection) collectionOnSession(sess *mgo.Session) *mgo.Collection {
	return sess.DB(c.Database).C(c.Name)
}

// FindID is a wrapper to the mgo FindId
func (c *Collection) FindID(id interface{}) *Query {
	q := &Query{
		MgoC: c.C(),
		MgoQ: c.C().FindId(id),
	}

	return q
}

// Find is a wrapper to the mgo Find. This is the entry point to the Query object.
// Populate makes a call to this method with a special meaning
func (c *Collection) Find(query interface{}) *Query {
	q := &Query{
		MgoC:     c.C(),
		Populate: false,
		Query:    nil,
	}

	if refactor, ok := query.(bson.M); ok {
		if refactor["$populate"] != nil {
			refactor["$and"] = refactor["$populate"]
			delete(refactor, "$populate")
			q.Populate = true
			query = refactor
		}
	}

	q.Query = query
	q.MgoQ = c.C().Find(query)

	return q
}

// FindID is convenience method for Collection.FindID
func FindID(doc Document, id interface{}) *Query {
	if d, ok := doc.(Document); ok {
		return d.GetColl().FindID(id)
	}

	return nil
}

// Find is convenience method for Collection.Find
func Find(doc Document, query interface{}) *Query {
	if d, ok := doc.(Document); ok {
		return d.GetColl().Find(query)
	}

	return nil
}

// Populate (TODO: make this wrapper)
func Populate(doc Document, query interface{}, ref string) {
}
