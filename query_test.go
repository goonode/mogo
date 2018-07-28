package bongo

import (
	"fmt"
	"testing"

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

			iter.NextPage(&results)
			So(len(results), ShouldEqual, iter.Pagination.OnPage)
			iter.NextPage(&results)
			So(len(results), ShouldEqual, iter.Pagination.OnPage)
			iter.NextPage(&results)
			So(len(results), ShouldEqual, iter.Pagination.OnPage)
			iter.NextPage(&results)
			So(len(results), ShouldEqual, iter.Pagination.OnPage)
		})

		Reset(func() {
			DBConn.Session.DB("bongotest").DropDatabase()
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
			DBConn.Session.DB("bongotest").DropDatabase()
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
			DBConn.Session.DB("bongotest").DropDatabase()
		})
	})
}
