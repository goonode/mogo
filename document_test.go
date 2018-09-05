package mogo

import (
	"testing"

	"github.com/globalsign/mgo"
	. "github.com/smartystreets/goconvey/convey"
)

func init() {
	_, _ = Connect(&Config{
		Database:         "mogotest",
		ConnectionString: "localhost",
	})
}

// BadDocument is not a valid document because it doesn't have
// the DocumentModel field
type BadDocument struct {
	Name    string
	Surname string
}

// DocumentWithModel is a valid document because it has the
// DocumentModel field and also define the collection name
type DocumentWithModel struct {
	DocumentModel `bson:",inline" collection:"test"`
	Name          string
	Surname       string
}

// DocumentWithModelAndIdx is a valid document because it has the
// DocumentModel field, define the collection name. Also it defines
// some index that will be stored in the collection
type DocumentWithModelAndIdx struct {
	DocumentModel `bson:",inline" collection:"test" idx:"{name,surname},unique"`
	Name          string `idx:"{name},unique,sparse"`
	Surname       string
}

type DocumentWithChildren struct {
	DocumentModel `bson:",inline" collection:"parent-collection" idx:"{name,surname},unique"`
	Name          string `idx:"{name},unique,sparse" collection:"parent-colleciton"` // WARN call is used outside DM
	Surname       string
	Childs        []RefField `ref:"DocumentChild"`
	Child         RefField   `ref:"DocumentChild"`
}

type DocumentWithChildrenNoRef struct {
	DocumentModel `bson:",inline" collection:"parent-collection" idx:"{name,surname},unique"`
	Name          string `idx:"{name},unique,sparse" collection:"parent-colleciton"`
	Surname       string
	Child         []RefField // This field should have ref tag
}
type DocumentChild struct {
	DocumentModel `bson:",inline" collection:"child-collection" idx:"{name,surname},unique"`
	Name          string `idx:"{name},unique,sparse"`
	Surname       string
}

func TestNewDocument(t *testing.T) {
	Convey("should create a new document if document is valid or panic if document is invalid", t, func() {
		doc := NewDoc(DocumentWithModelAndIdx{
			Name:    "MyName",
			Surname: "MySurname",
		}).(*DocumentWithModelAndIdx)

		So(doc.Name, ShouldEqual, "MyName")
		So(doc.Surname, ShouldEqual, "MySurname")

		So(func() { _ = NewDoc(BadDocument{}) }, ShouldPanic)
		So(func() { _ = NewDoc(DocumentWithModel{}).(*DocumentWithModel) }, ShouldNotPanic)
	})

	Convey("should create a new document passing a slice", t, func() {
		s := []*DocumentWithModelAndIdx{}
		doc := NewDoc(s).(*DocumentWithModelAndIdx)

		So(doc.Name, ShouldEqual, "")
		So(doc.Surname, ShouldEqual, "")
	})
}

func TestNewDocumentWithChildren(t *testing.T) {
	Convey("should create a new document if document is valid or panic if document is invalid", t, func() {
		So(func() {
			ModelRegistry.Register(DocumentWithChildren{},
				DocumentWithChildrenNoRef{},
				DocumentChild{})
		}, ShouldPanic)

		doc := NewDoc(DocumentWithChildren{
			Name:    "MyName",
			Surname: "MySurname",
		}).(*DocumentWithChildren)

		So(doc.Name, ShouldEqual, "MyName")
		So(doc.Surname, ShouldEqual, "MySurname")

		So(func() { _ = NewDoc(BadDocument{}) }, ShouldPanic)
		So(func() { _ = NewDoc(DocumentWithModel{}).(*DocumentWithModel) }, ShouldNotPanic)

		So(func() {
			_ = NewDoc(DocumentWithChildrenNoRef{
				Name:    "MyName",
				Surname: "MySurname",
			}).(*DocumentWithChildrenNoRef)
		}, ShouldPanic)
	})
}

func TestGetParsedIndex(t *testing.T) {
	ModelRegistry.Register(DocumentWithModelAndIdx{})
	Convey("should return the parsed indexes as defined in idx tag", t, func() {
		doc := NewDoc(DocumentWithModelAndIdx{}).(*DocumentWithModelAndIdx)
		pi := doc.GetParsedIndex("Name")
		So(pi, ShouldResemble, []ParsedIndex{
			ParsedIndex{[]string{"name"}, []string{"unique", "sparse"}, 0, false}})
		pi = doc.GetParsedIndex("Boh")
		So(pi, ShouldBeNil)
		rm := make(map[string][]ParsedIndex, 0)
		rm["DocumentModel"] = []ParsedIndex{ParsedIndex{[]string{"name", "surname"}, []string{"unique"}, 0, false}}
		rm["Name"] = []ParsedIndex{ParsedIndex{[]string{"name"}, []string{"unique", "sparse"}, 0, false}}
		rm["Surname"] = nil
		mi := doc.GetAllParsedIndex()
		So(mi, ShouldResemble, rm)
	})
}

func TestGetIndex(t *testing.T) {
	Convey("should return a  []*mgo.Index from the []ParsedIndex built from idx tag of the Name field", t, func() {
		doc := NewDoc(DocumentWithModelAndIdx{}).(*DocumentWithModelAndIdx)
		idx := doc.GetIndex("Name")
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
		doc := NewDoc(&DocumentWithModelAndIdx{
			Name: "MyFirst",
		}).(*DocumentWithModelAndIdx)
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

func TestDocumentSave(t *testing.T) {
	conn := getConnection()
	defer conn.Session.Close()

	ModelRegistry.Register(noHookDocument{}, hookedDocument{})

	Convey("Saving", t, func() {
		Convey("should be able to save a document with no hooks, update id, and use new tracker", func() {
			doc := NewDoc(noHookDocument{}).(*noHookDocument)
			doc.Name = "foo"
			So(doc.IsNew(), ShouldEqual, true)
			err := doc.Save()
			So(err, ShouldEqual, nil)
			So(doc.ID.Valid(), ShouldEqual, true)
			So(doc.IsNew(), ShouldEqual, false)
		})

		Convey("should be able to save a document with save hooks", func() {
			doc := NewDoc(hookedDocument{}).(*hookedDocument)
			err := doc.Save()

			So(err, ShouldEqual, nil)
			So(doc.RanBeforeSave, ShouldEqual, true)
			So(doc.RanAfterSave, ShouldEqual, true)
		})

		Convey("should return a validation error if the validate method has things in the return value", func() {
			doc := NewDoc(validatedDocument{}).(*validatedDocument)
			err := doc.Save()

			v, ok := err.(*ValidationError)
			So(ok, ShouldEqual, true)
			So(v.Errors[0].Error(), ShouldEqual, "test validation error")
		})

		Convey("should be able to save an existing document", func() {
			doc := NewDoc(noHookDocument{}).(*noHookDocument)
			doc.Name = "foo"
			So(doc.IsNew(), ShouldEqual, true)

			err := doc.Save()
			So(err, ShouldEqual, nil)
			So(doc.ID.Valid(), ShouldEqual, true)
			So(doc.IsNew(), ShouldEqual, false)

			err = doc.Save()

			So(err, ShouldEqual, nil)
			count, err := doc.GetColl().C().Count()
			So(err, ShouldEqual, nil)
			So(count, ShouldEqual, 1)
		})

		Convey("should set created and modified dates", func() {

			doc := NewDoc(noHookDocument{}).(*noHookDocument)
			doc.Name = "foo"

			err := doc.Save()
			So(err, ShouldEqual, nil)
			So(doc.Created.UnixNano(), ShouldEqual, doc.GetModified().UnixNano())

			err = doc.Save()
			So(err, ShouldEqual, nil)
			So(doc.Modified.UnixNano(), ShouldBeGreaterThan, doc.GetCreated().UnixNano())
		})

		Reset(func() {
			conn.Session.DB("mogotest").DropDatabase()
		})
	})
}
