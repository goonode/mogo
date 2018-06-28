package bongo

import (
	"fmt"
	"testing"

	"github.com/globalsign/mgo"
	. "github.com/smartystreets/goconvey/convey"
)

type HomeAddress struct {
	Street string
	Suite  string
	City   string
	State  string
	Zip    string
}

type Person struct {
	DocumentModel `bson:",inline" coll:"persons"`
	FirstName     string `idx:"{firstname},unique"`
	LastName      string `idx:"{lastname},unique"`
	Gender        string
	HomeAddress   HomeAddress `idx:"{homeaddress.street, homeaddress.city},unique,sparse"`
}

func TestSave(t *testing.T) {
	conn, _ := Connect(&Config{
		ConnectionString: "localhost",
		Database:         "bongotest",
	})
	conn.Context.Set("foo", "bar")
	defer DBConn.Session.Close()

	ModelRegistry.Register(noHookDocumentWithSlice{}, hookedDocument{})

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

func TestSaveWithChildStruct(t *testing.T) {
	conn, _ := Connect(&Config{
		ConnectionString: "localhost",
		Database:         "bongotest",
	})
	conn.Context.Set("foo", "bar")
	defer DBConn.Session.Close()

	ModelRegistry.Register(Person{})

	Convey("save using direct and wrapper methods", t, func() {
		Convey("should be able to save or update", func() {
			d := NewDocument(Person{
				FirstName: "Bingo",
				LastName:  "Bongo",
			}).(*Person)
			d.HomeAddress.Street = "Main"
			err := d.GetColl().Save(d)

			d = NewDocument(Person{
				FirstName: "Bongo",
				LastName:  "Bingo",
			}).(*Person)
			d.HomeAddress.Street = "Main" // Unique index violation
			err = d.GetColl().Save(d)
			So(err, ShouldNotBeNil)
			So(err.(*mgo.LastError).Code, ShouldEqual, 11000)
		})

		Reset(func() {
			conn.Session.DB("bongotest").DropDatabase()
		})
	})
}
