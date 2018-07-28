package bongo

import (
	"math"
	"reflect"

	"github.com/globalsign/mgo"
)

// Query is the mgo.Query wrapper
type Query struct {
	MgoC *mgo.Collection
	MgoQ *mgo.Query

	Pagination *Paginate
}

// Iter is the mgo.Iter wrapper
type Iter struct {
	MgoQ    *mgo.Query
	MgoI    *mgo.Iter
	Timeout bool
	Err     error

	Pagination *Paginate
}

// Paginate ...
type Paginate struct {
	Page   int `json:"page"`   // Current page
	Pages  int `json:"pages"`  // Total pages
	N      int `json:"items"`  // Items per pages
	T      int `json:"total"`  // Total records in query
	OnPage int `json:"onPage"` // Records in current page
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
		MgoQ:       q.MgoQ,
		MgoI:       q.MgoQ.Iter(),
		Pagination: q.Pagination,
		Timeout:    false,
		Err:        nil,
	}

	return i
}

// Limit is a wrapper around mgo.Query.Limit
func (q *Query) Limit(n int) *Query {
	q.MgoQ = q.MgoQ.Limit(n)
	return q
}

// Skip is a wrapper around mgo.Query.Skip
func (q *Query) Skip(n int) *Query {
	q.MgoQ = q.MgoQ.Skip(n)
	return q
}

// Paginate prepares the Query to allow pagination
func (q *Query) Paginate(n int) *Query {
	q.Pagination = &Paginate{
		Page:   0,
		Pages:  0,
		OnPage: 0,
		T:      0,
		N:      n,
	}

	return q
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

// Next is a wrapper around mgo.Iter.Next. It executes AfterFindHook and the updates
// the NewTracker interface if needed.
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
		if i.MgoI.Timeout() {
			i.Timeout = true
			return false
		}

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

// NextPage is the paginated version of the Next iterator. It fills
// the results slice using the Pagination field of the Iterator.
// Before using this the Query should be initialized using the Paginate()
// receiver.
func (i *Iter) NextPage(results interface{}) bool {
	var n int
	var err error

	rv := reflect.ValueOf(results)
	if rv.Kind() != reflect.Ptr {
		panic("results argument must be a slice")
	}

	if i.Pagination == nil {
		return false
	}

	// if len(results) < i.Pagination.N {
	// 	panic(fmt.Sprintf("passed slice size (%d) is lower than paginate size (%d)", len(results), i.Pagination.N))
	// }

	if i.Pagination.T == 0 {
		n, err = i.MgoQ.Count()
		if err != nil {
			i.Err = err
		}

		i.Pagination.T = n
		i.Pagination.Page = 0
		i.Pagination.Pages = int(math.Ceil(float64(n) / float64(i.Pagination.N)))
	}

	if i.Pagination.Page >= i.Pagination.Pages {
		i.Pagination.Page = 1
	} else {
		i.Pagination.Page++
	}
	i.MgoQ = i.MgoQ.Skip((i.Pagination.Page - 1) * i.Pagination.N).Limit(i.Pagination.N)
	i.MgoI = i.MgoQ.Iter()

	r := NewDoc(results)
	sv := rv.Elem()
	sv = sv.Slice(0, sv.Cap())
	l := 0
	// TODO: error management
	for i.Next(r) {
		if sv.Len() == l {
			sv = reflect.Append(sv, reflect.ValueOf(r))
			sv = sv.Slice(0, sv.Cap())
		} else {
			sv.Index(l).Set(reflect.ValueOf(r))
		}
		r = NewDoc(results)
		l++
	}

	rv.Elem().Set(sv.Slice(0, l))
	i.Pagination.OnPage = l

	if i.Pagination.Page == i.Pagination.Pages {
		return false
	}

	return true
}

// Done is a wrapper around mgo.Iter.Done
func (i *Iter) Done() bool {
	return i.MgoI.Done()
}
