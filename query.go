package bongo

import (
	"github.com/globalsign/mgo"
)

// Query is the mgo.Query wrapper
type Query struct {
	MgoC *mgo.Collection
	MgoQ *mgo.Query
}

// Iter is the mgo.Iter wrapper
type Iter struct {
	MgoI    *mgo.Iter
	Timeout bool
	Err     error
}

// C direct access to mgo driver Collection layer
func (q *Query) C() *mgo.Collection {
	return q.MgoC
}

// Q direct access to mgo driver Query layer
func (q *Query) Q() *mgo.Query {
	return q.MgoQ
}

// All is a wrapper around mgo.Query.All
func (q *Query) All(result interface{}) error {
	return nil
}

// Iter is a wrapper around mgo.Query.Iter
func (q *Query) Iter() *Iter {
	i := &Iter{
		MgoI: q.MgoQ.Iter(),
	}

	return i
}

// Limit is a wrapper around mgo.Query.Limit
func (q *Query) Limit(n int) *Query {
	return nil
}

// One is a wrapper around mgo.Query.One
func (q *Query) One(result interface{}) error {
	var iname string
	var err error
	var ok bool
	var d Model

	if d, ok = result.(Document); ok {
		iname, _ = d.GetMe()
	} else {
		panic("result is not a bongo document")
	}

	if err = q.MgoQ.One(result); err != nil {
		d.SetMe(iname, result)
		return err
	}
	// Restoring the iname Document field
	d.SetMe(iname, result)

	if hook, ok := d.(AfterFindHook); ok {
		err = hook.AfterFind()
		if err != nil {
			return err
		}
	}

	// We retrieved it, so set new to false
	if newt, ok := d.(NewTracker); ok {
		newt.SetIsNew(false)
	}

	return nil
}

// Next is a wrapper around mgo.Iter.Next
func (i *Iter) Next(result interface{}) bool {
	var iname string
	var err error
	var ok bool
	var d Model

	if d, ok = result.(Document); ok {
		iname, _ = d.GetMe()
	} else {
		panic("result is not a bongo document")
	}

	if ok = i.MgoI.Next(result); !ok {
		i.Err = i.MgoI.Err()
		d.SetMe(iname, result)
		return false
	}

	d.SetMe(iname, result)
	if hook, ok := d.(AfterFindHook); ok {
		err = hook.AfterFind()
		if err != nil {
			i.Err = err
			return false
		}
	}

	// We retrieved it, so set new to false
	if newt, ok := d.(NewTracker); ok {
		newt.SetIsNew(false)
	}
	return true
}
