package mogo

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestDelete(t *testing.T) {
	conn, _ := Connect(&Config{
		ConnectionString: "localhost",
		Database:         "mogotest",
	})
	conn.Context.Set("foo", "bar")
	defer DBConn.Session.Close()

	ModelRegistry.Register(noHookDocumentWithSlice{}, hookedDocument{})

	Convey("delete using direct and/or wrapper methods", t, func() {
		Convey("should be able to delete by id", func() {
			d := NewDoc(hookedDocument{}).(*hookedDocument)
			err := d.GetColl().Save(d)
			So(err, ShouldBeNil)
			err = Remove(d)
			So(err, ShouldBeNil)
			// id := d.GetId()

			e := NewDoc(hookedDocument{}).(*hookedDocument)
			err = FindID(e, d.BsonID()).One(e)
			So(err.Error(), ShouldEqual, "not found")
		})

		Reset(func() {
			conn.Session.DB("mogotest").DropDatabase()
		})
	})
}
