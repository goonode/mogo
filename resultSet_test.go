package bongo

import (
	"fmt"
	"testing"

	"github.com/globalsign/mgo/bson"
	. "github.com/smartystreets/goconvey/convey"
)

func TestResultSet(t *testing.T) {
	conn := getConnection()
	doc := NewDocument(noHookDocument{}, conn).(*noHookDocument)
	// collection := conn.Collection("tests")
	defer doc.connection.Session.Close()

	Convey("Basic find/pagination", t, func() {
		// Create 10 things
		for i := 0; i < 10; i++ {
			doc.Name = fmt.Sprintf("Number_%d", i)
			Save(doc)
			doc.MakeAsNew()
		}

		Convey("should let you iterate through all results without paginating", func() {
			rset := Find(doc, nil)
			defer rset.Free()
			count := 0

			// doc := NewDocumentModel(noHookDocument{}, conn).(*noHookDocument)

			for rset.Next(doc) {
				count++
			}

			So(count, ShouldEqual, 10)
		})

		Convey("should let you paginate and get pagination info", func() {
			rset := Find(doc, nil)
			defer rset.Free()
			info, err := rset.Paginate(3, 1)
			So(err, ShouldEqual, nil)
			So(info.TotalRecords, ShouldEqual, 10)
			So(info.TotalPages, ShouldEqual, 4)
			So(info.Current, ShouldEqual, 1)
			So(info.PerPage, ShouldEqual, 3)
			So(info.RecordsOnPage, ShouldEqual, 3)

			rset2 := Find(doc, nil)
			defer rset2.Free()
			info, err = rset2.Paginate(3, 4)
			So(err, ShouldEqual, nil)
			So(info.TotalRecords, ShouldEqual, 10)
			So(info.TotalPages, ShouldEqual, 4)
			So(info.Current, ShouldEqual, 4)
			So(info.PerPage, ShouldEqual, 3)
			So(info.RecordsOnPage, ShouldEqual, 1)
		})

		Reset(func() {
			conn.Session.DB("bongotest").DropDatabase()
		})
	})

	Convey("Find/pagination w/ query", t, func() {
		// Create 10 things
		for i := 0; i < 5; i++ {
			doc := NewDocument(noHookDocument{}, conn).(*noHookDocument)
			doc.Name = "foo"
			Save(doc)
		}
		for i := 0; i < 5; i++ {
			doc.Name = "bar"
			Save(doc)
		}

		Convey("should let you iterate through all filtered results without paginating", func() {
			rset := Find(doc, bson.M{
				"name": "foo",
			})
			defer rset.Free()

			count := 0

			// doc := &noHookDocument{}

			for rset.Next(doc) {
				count++
			}

			So(count, ShouldEqual, 5)
		})

		Convey("should let you paginate and get pagination info on filtered query", func() {
			rset := Find(doc, bson.M{
				"name": "foo",
			})
			defer rset.Free()
			info, err := rset.Paginate(3, 1)
			So(err, ShouldEqual, nil)
			So(info.TotalRecords, ShouldEqual, 5)
			So(info.TotalPages, ShouldEqual, 2)
			So(info.Current, ShouldEqual, 1)
			So(info.PerPage, ShouldEqual, 3)
			So(info.RecordsOnPage, ShouldEqual, 3)

			rset2 := Find(doc, bson.M{
				"name": "foo",
			})
			defer rset2.Free()
			info, err = rset2.Paginate(3, 2)
			So(err, ShouldEqual, nil)
			So(info.TotalRecords, ShouldEqual, 5)
			So(info.TotalPages, ShouldEqual, 2)
			So(info.Current, ShouldEqual, 2)
			So(info.PerPage, ShouldEqual, 3)
			So(info.RecordsOnPage, ShouldEqual, 2)
		})

		Reset(func() {
			conn.Session.DB("bongotest").DropDatabase()
		})
	})

	Convey("hooks", t, func() {
		// Create 10 things
		for i := 0; i < 10; i++ {
			doc := NewDocument(noHookDocument{}, conn).(*noHookDocument)
			Save(doc)
		}

		Convey("should let you iterate through all results without paginating", func() {
			rset := Find(doc, nil)
			defer rset.Free()
			count := 0

			doc := NewDocument(hookedDocument{}, conn).(*hookedDocument)

			for rset.Next(doc) {
				So(doc.RanAfterFind, ShouldEqual, true)
				count++
			}

			So(count, ShouldEqual, 10)
		})

		Reset(func() {
			conn.Session.DB("bongotest").DropDatabase()
		})
	})
}
