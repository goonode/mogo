package bongo

import (
	"fmt"
	"reflect"
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

// Find is a wrapper to the mgo Find
func (c *Collection) Find(query interface{}) *Query {
	q := &Query{
		MgoC: c.C(),
		MgoQ: c.C().Find(query),
	}

	return q
}

// Populate ... TODO:
func (c *Collection) Populate(doc Document, query interface{}, ref string) *Query {
	i, r := ModelRegistry.SearchRef(doc, ref)
	v := ModelRegistry.New(ref)

	if v != nil {
		// Gathering internal infos
		switch i.Type.Field(r.Idx).Type.Kind() {
		case reflect.Slice:
			fallthrough
		case reflect.Map:
			fallthrough
		default:
			f := ValueOf(doc).Field(r.Idx).Interface().(RefField)
			q := bson.M{"_id": f.ID}
			fmt.Println(q, doc.GetColl().C().FullName)
		}
	}

	return nil
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

// Populate ...
func Populate(doc Document, query interface{}, ref string) {
}
