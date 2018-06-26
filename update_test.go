package bongo

import (
	"fmt"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestSave(t *testing.T) {
	conn, _ := Connect(&Config{
		ConnectionString: "localhost",
		Database:         "bongotest",
	})
	conn.Context.Set("foo", "bar")
	defer DBConn.Session.Close()

	modelRegistry.Register(noHookDocumentWithSlice{}, hookedDocument{})

	Convey("save using direct and wrapper methods", t, func() {
		Convey("should be able to save or update", func() {
			d1 := NewDocument(hookedDocument{}).(*hookedDocument)
			err := d1.GetColl().Save(d1)
			So(err, ShouldBeNil)

			d2 := NewDocument(hookedDocument{
				Name: "Olo",
			}).(*hookedDocument)
			err = d2.GetColl().Save(d2)
			So(err, ShouldBeNil)
			d2f := NewDocument(hookedDocument{}).(*hookedDocument)
			err = FindByID(d2f, d2.ID)
			So(err, ShouldBeNil)
			So(d2f.ID, ShouldEqual, d2.ID)
			d2.Name = "olO"
			err = Save(d2)
			So(err, ShouldBeNil)
			So(d2.Name, ShouldNotEqual, d2f.Name)

			d3 := NewDocument(hookedDocument{}).(*hookedDocument)
			for i := 0; i < 9; i++ {
				d3.MakeAsNew()
				d3.Name = fmt.Sprintf("%d_Olo", i)
				d3.Surname = fmt.Sprintf("olO_%d", i)
				err = Save(d3)
				So(err, ShouldBeNil)
			}
		})

		Reset(func() {
			conn.Session.DB("bongotest").DropDatabase()
		})
	})
}
