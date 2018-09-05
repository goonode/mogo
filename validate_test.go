package mogo

import (
	"testing"

	"github.com/globalsign/mgo/bson"
	. "github.com/smartystreets/goconvey/convey"
)

func TestValidation(t *testing.T) {
	connection := getConnection()
	defer connection.Session.Close()

	ModelRegistry.Register(noHookDocument{}, hookedDocument{})

	Convey("Validation", t, func() {
		Convey("ValidateRequired()", func() {
			So(ValidateRequired("foo"), ShouldEqual, true)
			So(ValidateRequired(""), ShouldEqual, false)
			So(ValidateRequired(0), ShouldEqual, false)
			So(ValidateRequired(1), ShouldEqual, true)
		})

		Convey("ValidateInclusionIn()", func() {
			So(ValidateInclusionIn("foo", []string{"foo", "bar", "baz"}), ShouldEqual, true)
			So(ValidateInclusionIn("bing", []string{"foo", "bar", "baz"}), ShouldEqual, false)
		})

		Convey("ValidateMongoIDRef()", func() {

			defer func() {
				connection.Session.DB("mogotest").DropDatabase()
			}()

			// Make the doc

			doc := NewDoc(&noHookDocument{}).(*noHookDocument)

			err := Save(doc)

			So(err, ShouldEqual, nil)
			So(ValidateMongoIDRef(doc.ID, doc.GetColl()), ShouldEqual, true)
			So(ValidateMongoIDRef(bson.NewObjectId(), doc.GetColl()), ShouldEqual, false)
			So(ValidateMongoIDRef(bson.NewObjectId(), connection.Collection("other_collection")), ShouldEqual, false)
		})
	})
}
