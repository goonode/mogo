package bongo

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestDelete(t *testing.T) {
	conn, _ := Connect(&Config{
		ConnectionString: "localhost",
		Database:         "bongotest",
	})
	conn.Context.Set("foo", "bar")
	defer DBConn.Session.Close()

	ModelRegistry.Register(noHookDocumentWithSlice{}, hookedDocument{})

	Convey("delete using direct and/or wrapper methods", t, func() {
		Convey("should be able to delete by id", func() {
			d := NewDocument(hookedDocument{}).(*hookedDocument)
			err := d.GetColl().Save(d)
			So(err, ShouldBeNil)
			err = DeleteDocument(d)
			So(err, ShouldBeNil)
			id := d.GetID()

			e := NewDocument(hookedDocument{}).(*hookedDocument)
			err = FindByID(e, id)
			So(err, ShouldResemble, &DocumentNotFoundError{})
		})

		Reset(func() {
			conn.Session.DB("bongotest").DropDatabase()
		})
	})
}
