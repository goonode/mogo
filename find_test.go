package bongo

import (
	"fmt"
	"testing"

	"github.com/globalsign/mgo/bson"

	. "github.com/smartystreets/goconvey/convey"
)

func TestFinding(t *testing.T) {
	conn, _ := Connect(&Config{
		ConnectionString: "localhost",
		Database:         "bongotest",
	})
	conn.Context.Set("foo", "bar")
	defer DBConn.Session.Close()

	modelRegistry.Register(noHookDocumentWithSlice{}, hookedDocument{})
	Convey("finding using direct methods", t, func() {
		Convey("should be able to find by id", func() {
			d1 := NewDocument(noHookDocumentWithSlice{}).(*noHookDocumentWithSlice)
			err := d1.GetColl().FindByID(bson.ObjectIdHex("5b2e8488c5285a422a9012ba"), d1)
			d2 := NewDocument(hookedDocument{
				Name: "Olo",
			}).(*hookedDocument)
			d2.GetColl().Save(d2)
			err = d2.GetColl().FindByID(d2.ID, d2)

			fmt.Println(err)
		})

		Reset(func() {
			conn.Session.DB("bongotest").DropDatabase()
		})
	})

	Convey("finding using wrapper methods", t, func() {
		Convey("should be able to find by id", func() {
			d1 := NewDocument(noHookDocumentWithSlice{}).(*noHookDocumentWithSlice)
			err := FindByID(d1, bson.ObjectIdHex("5b2e8488c5285a422a9012ba"))
			d2 := NewDocument(hookedDocument{
				Name: "Olo",
			}).(*hookedDocument)
			d2.GetColl().Save(d2)
			err = FindByID(d2, d2.ID)
			fmt.Println(d2.ID.Hex(), err)
		})
		Reset(func() {
			conn.Session.DB("bongotest").DropDatabase()
		})
	})

}
