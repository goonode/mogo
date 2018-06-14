package bongo

import (
	"testing"

	"github.com/globalsign/mgo"
	. "github.com/smartystreets/goconvey/convey"
)

// BadDocument is not a valid document because it doesn't have
// the DocumentModel field
type BadDocument struct {
	Name    string
	Surname string
}

// DocumentWithModel is a valid document because it has the
// DocumentModel field and also define the collection name
type DocumentWithModel struct {
	DocumentModel `bson:",inline" coll:"test"`
	Name          string
	Surname       string
}

// DocumentWithModelAndIdx is a valid document because it has the
// DocumentModel field, define the collection name. Also it defines
// some index that will be stored in the collection
type DocumentWithModelAndIdx struct {
	DocumentModel `bson:",inline" coll:"test" idx:"{name,surname},unique"`
	Name          string `idx:"{name},unique,sparse"`
	Surname       string
}

func TestNewDocument(t *testing.T) {
	Convey("should create a new document if document is valid or panic if document is invalid", t, func() {
		So(func() { _ = NewDocumentModel(BadDocument{}) }, ShouldPanic)
		So(func() { _ = NewDocumentModel(DocumentWithModel{}).(*DocumentWithModel) }, ShouldNotPanic)
	})
}

func TestGetParsedIndex(t *testing.T) {
	Convey("should return the parsed indexes as defined in idx tag", t, func() {
		doc := NewDocumentModel(DocumentWithModelAndIdx{}).(*DocumentWithModelAndIdx)
		pi := doc.GetParsedIndex("_Name")
		So(pi, ShouldResemble, []ParsedIndex{
			ParsedIndex{[]string{"name"}, []string{"unique", "sparse"}}})
		pi = doc.GetParsedIndex("Boh")
		So(pi, ShouldBeNil)
		rm := make(map[string][]ParsedIndex, 0)
		rm["DocumentModel"] = []ParsedIndex{ParsedIndex{[]string{"name", "surname"}, []string{"unique"}}}
		rm["_Name"] = []ParsedIndex{ParsedIndex{[]string{"name"}, []string{"unique", "sparse"}}}
		rm["_Surname"] = nil
		mi := doc.GetAllParsedIndex()
		So(mi, ShouldResemble, rm)
	})
}

func TestGetIndex(t *testing.T) {
	Convey("should return a  []*mgo.Index from the []ParsedIndex built from idx tag of the Name field", t, func() {
		doc := NewDocumentModel(DocumentWithModelAndIdx{}).(*DocumentWithModelAndIdx)
		idx := doc.GetIndex("_Name")
		So(len(idx), ShouldBeGreaterThan, 0)
		mi := &mgo.Index{
			Key:    []string{"name"},
			Unique: true,
			Sparse: true,
		}
		So(idx[0], ShouldResemble, mi)
	})
}

func TestGetAllIndex(t *testing.T) {
	Convey("should return a []*mgo.Index from the []ParsedIndex built from idx tags of all fields", t, func() {
		doc := NewDocumentModel(DocumentWithModelAndIdx{}).(*DocumentWithModelAndIdx)
		idx := doc.GetAllIndex()
		So(len(idx), ShouldBeGreaterThan, 0)
		mi := &mgo.Index{
			Key:    []string{"name"},
			Unique: true,
			Sparse: true,
		}
		So(idx[1], ShouldResemble, mi)
	})
}
