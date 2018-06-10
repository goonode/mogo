package bongo

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

type DocumentWithTags struct {
	DocumentModel `bson:",inline" coll:"test" idx:"{name,surname},unique,sparse,{surname},unique"`
	Name          string
	Surname       string
}

func TestGetCollectionName(t *testing.T) {
	Convey("should return the collection name defined in tag", t, func() {
		d := &DocumentWithTags{
			Name:    "bongo",
			Surname: "bongo",
		}

		d.GetIndexedFields(d)
		So(d.GetCollectionName(d), ShouldEqual, "test")
	})
}
