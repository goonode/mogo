package bongo

import (
	"time"

	"github.com/globalsign/mgo"

	"strings"
)

// BeforeSaveHook ...
type BeforeSaveHook interface {
	BeforeSave(*Collection) error
}

// AfterSaveHook ...
type AfterSaveHook interface {
	AfterSave(*Collection) error
}

// BeforeDeleteHook ...
type BeforeDeleteHook interface {
	BeforeDelete(*Collection) error
}

// AfterDeleteHook ...
type AfterDeleteHook interface {
	AfterDelete(*Collection) error
}

// AfterFindHook ...
type AfterFindHook interface {
	AfterFind(*Collection) error
}

// ValidateHook ...
type ValidateHook interface {
	Validate(*Collection) []error
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

// Collection ...
type Collection struct {
	Name       string
	Database   string
	Context    *Context
	Connection *Connection
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

// Collection ...
func (c *Collection) Collection() *mgo.Collection {
	return c.Connection.Session.DB(c.Database).C(c.Name)
}

// collectionOnSession ...
func (c *Collection) collectionOnSession(sess *mgo.Session) *mgo.Collection {
	return sess.DB(c.Database).C(c.Name)
}
