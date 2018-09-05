package mogo

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/globalsign/mgo/bson"

	. "github.com/smartystreets/goconvey/convey"
)

func init() {
}

func TestQuery(t *testing.T) {
	conn := getConnection()
	defer conn.Session.Close()

	ModelRegistry.Register(noHookDocument{}, hookedDocument{})

	doc := NewDoc(noHookDocument{}).(*noHookDocument)
	defer DBConn.Session.Close()

	Convey("Basic find/pagination", t, func() {
		// Create 10 things
		for i := 0; i < 10; i++ {
			doc.Name = fmt.Sprintf("Number_%d", i)
			Save(doc)
			doc.AsNew()
		}

		Convey("should let you iterate through all results without paginating", func() {
			count := 0
			iter := doc.Find(nil).Iter()

			for iter.Next(doc) {
				count++
			}
			So(count, ShouldEqual, 10)
		})

		Convey("should let you paginate and get pagination info", func() {
			iter := doc.Find(nil).Paginate(3).Iter()
			results := make([]*noHookDocument, 3)

			for iter.NextPage(&results) {
				So(len(results), ShouldEqual, iter.Pagination.OnPage)
			}
		})

		Reset(func() {
			DBConn.Session.DB("mogotest").DropDatabase()
		})
	})

	Convey("Find/pagination w/ query", t, func() {
		// Create 10 things
		for i := 0; i < 5; i++ {
			doc := NewDoc(noHookDocument{}).(*noHookDocument)
			doc.Name = "foo"
			Save(doc)
		}
		for i := 0; i < 5; i++ {
			doc.Name = "bar"
			Save(doc)
		}

		Convey("should let you iterate through all filtered results without paginating", func() {
		})

		Convey("should let you paginate and get pagination info on filtered query", func() {
		})

		Reset(func() {
			DBConn.Session.DB("mogotest").DropDatabase()
		})
	})

	Convey("hooks", t, func() {
		// Create 10 things
		for i := 0; i < 10; i++ {
			doc := NewDoc(hookedDocument{}).(*hookedDocument)
			Save(doc)
		}

		Convey("should let you iterate through all results without paginating", func() {
		})

		Reset(func() {
			DBConn.Session.DB("mogotest").DropDatabase()
		})
	})
}

func scrach(d DocumentModel, f string) *Query {
	// Populate analisys:
	// 	we need to populate a RefField or RefFieldSlice or by calling Populate on DocumentModel
	//	To do this we need to know the collection to operate on.
	//	We could use the modelregistry to get the reference to the object which contains
	//	the model;

	_, i, _ := ModelRegistry.Exists(d.me)
	if i == nil { // model is not registered
		return nil
	}

	r := i.Refs[f]
	if !r.Exists { // Field name not exists
		return nil
	}

	iField := reflect.ValueOf(d.me).Elem().Field(r.Idx).Interface()
	t := ModelRegistry.New(r.Ref).(Document)
	var q = bson.M{"$populate": make([]bson.M, 0)}

	switch i.Refs[f].Kind {
	case reflect.Slice:
		var inner = bson.M{"$or": make([]bson.M, 0)}

		field := iField.(RefFieldSlice)
		for i := range field {
			inner["$or"] = append(inner["$or"].([]bson.M), bson.M{"_id": field[i].ID})
		}
		q["$populate"] = append(q["$populate"].([]bson.M), inner)
		return Find(t, q)
	default:
		field := iField.(RefField)
		var inner = bson.M{"_id": field.ID}

		q["$populate"] = append(q["$populate"].([]bson.M), inner)
		return Find(t, q)
	}
}

func TestPopulate(t *testing.T) {

	ModelRegistry.Register(Bongo{}, Macao{})

	Convey("Populate", t, func() {
		mogo := NewDoc(Bongo{}).(*Bongo)

		// Give some friends to mogo
		for i := 0; i < 10; i++ {
			macao := NewDoc(Macao{}).(*Macao)
			macao.Name = fmt.Sprintf("Macky%d", i)
			Save(macao)
			mogo.Friends = append(mogo.Friends, &RefField{ID: macao.ID})
		}
		// But the mogo best friend is
		macao := NewDoc(Macao{}).(*Macao)
		macao.Name = "Polly"
		Save(macao)
		mogo.BestFriend = RefField{ID: macao.ID}
		Save(mogo)

		// mogo.BestFriend.Populate(nil)
		Convey("Build a populate query and get results", func() {
			// Trying populate ...
			q := mogo.Populate("Friends")

			result := make([]Macao, 0)
			q.All(&result)
			for i := range result {
				fmt.Println(result[i])
			}
		})

		Convey("Build a populate query, add a filter and get results", func() {
			// Trying populate ...
			q := mogo.Populate("Friends").Find(bson.M{"name": "Macky2"})

			result := make([]Macao, 0)
			q.All(&result)
			for i := range result {
				fmt.Println(result[i])
			}
		})

		Reset(func() {
			DBConn.Session.DB("mogotest").DropDatabase()
		})
	})
}
