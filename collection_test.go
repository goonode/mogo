package mogo

import (
	"errors"
	"testing"

	"github.com/globalsign/mgo/bson"
	. "github.com/smartystreets/goconvey/convey"
)

type extraData struct {
	SubColors []string
}

type noHookDocument struct {
	DocumentModel `bson:",inline" coll:"nohooked-test" idx:"{name},unique,sparse"`
	Name          string
}

type noHookDocumentWithSlice struct {
	DocumentModel `bson:",inline" coll:"nohooked-test" idx:"{name},unique,sparse"`
	Name          string
	Colors        []string
	ColorMap      map[string][]extraData
	SubColor      extraData
	SubColors     []extraData
}

type hookedDocument struct {
	DocumentModel   `bson:",inline" coll:"hooked-test"`
	Name            string `idx:"{name,surname},unique"`
	Surname         string
	RanBeforeSave   bool
	RanAfterSave    bool
	RanBeforeDelete bool
	RanAfterDelete  bool
	RanAfterFind    bool
}

func (h *hookedDocument) BeforeSave() error {
	h.RanBeforeSave = true
	c := h.GetColl()
	So(c.Context.Get("foo"), ShouldEqual, "bar")
	return nil
}

func (h *hookedDocument) AfterSave() error {
	h.RanAfterSave = true
	c := h.GetColl()
	So(c.Context.Get("foo"), ShouldEqual, "bar")
	return nil
}

func (h *hookedDocument) BeforeDelete() error {
	h.RanBeforeDelete = true
	c := h.GetColl()
	So(c.Context.Get("foo"), ShouldEqual, "bar")
	return nil
}

func (h *hookedDocument) AfterDelete() error {
	h.RanAfterDelete = true
	c := h.GetColl()
	So(c.Context.Get("foo"), ShouldEqual, "bar")
	return nil
}

func (h *hookedDocument) AfterFind() error {
	h.RanAfterFind = true
	c := h.GetColl()
	So(c.Context.Get("foo"), ShouldEqual, "bar")
	return nil
}

type validatedDocument struct {
	DocumentModel `bson:",inline" coll:"validated-collection"`
	Name          string
}

func (v *validatedDocument) Validate() []error {
	return []error{errors.New("test validation error")}
}

func TestCollection(t *testing.T) {
	conn := getConnection()
	defer conn.Session.Close()

	ModelRegistry.Register(noHookDocument{}, hookedDocument{})

	Convey("Saving", t, func() {
		Convey("should be able to save a document with no hooks, update id, and use new tracker", func() {
			doc := NewDoc(noHookDocument{}).(*noHookDocument)
			doc.Name = "foo"
			So(doc.IsNew(), ShouldEqual, true)

			// err := Save(&doc)
			err := doc.Save()
			So(err, ShouldEqual, nil)
			So(doc.ID.Valid(), ShouldEqual, true)
			So(doc.IsNew(), ShouldEqual, false)
		})

		Convey("should be able to save a document with save hooks", func() {
			doc := NewDoc(hookedDocument{}).(*hookedDocument)
			err := Save(doc)

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

			err = Save(doc)

			So(err, ShouldEqual, nil)
			count, err := doc.GetColl().C().Count()
			So(err, ShouldEqual, nil)
			So(count, ShouldEqual, 1)
		})

		Convey("should set created and modified dates", func() {

			doc := NewDoc(noHookDocument{}).(*noHookDocument)
			doc.Name = "foo"

			err := Save(doc)
			So(err, ShouldEqual, nil)
			So(doc.Created.UnixNano(), ShouldEqual, doc.GetModified().UnixNano())

			err = Save(doc)
			So(err, ShouldEqual, nil)
			So(doc.Modified.UnixNano(), ShouldBeGreaterThan, doc.GetCreated().UnixNano())
		})

		Reset(func() {
			conn.Session.DB("mogotest").DropDatabase()
		})
	})

	Convey("Find by ID", t, func() {
		doc := NewDoc(noHookDocument{}).(*noHookDocument)
		err := Save(doc)
		So(err, ShouldEqual, nil)

		d2 := NewDoc(hookedDocument{}).(*hookedDocument)
		err = Save(d2)
		So(err, ShouldEqual, nil)

		Convey("should find a doc by id", func() {
			newDoc := NewDoc(noHookDocument{}).(*noHookDocument)
			err := newDoc.FindID(doc.GetID()).One(newDoc)
			So(err, ShouldEqual, nil)
			So(newDoc.ID.Hex(), ShouldEqual, doc.ID.Hex())
		})

		Convey("should find a doc by id and run afterFind", func() {
			newDoc := NewDoc(hookedDocument{}).(*hookedDocument)
			err := newDoc.FindByID(d2.GetID(), newDoc)
			So(err, ShouldEqual, nil)
			So(newDoc.ID.Hex(), ShouldEqual, d2.ID.Hex())
			So(newDoc.RanAfterFind, ShouldEqual, true)
		})

		Convey("should return a document not found error if doc not found", func() {
			var newDoc = NewDoc(&noHookDocument{}).(*noHookDocument)
			err := FindID(newDoc, bson.NewObjectId()).One(newDoc)
			So(err.Error(), ShouldEqual, "not found")
		})

		Reset(func() {
			conn.Session.DB("mogotest").DropDatabase()
		})
	})

	Convey("FindOne", t, func() {
		doc := NewDoc(hookedDocument{}).(*hookedDocument)
		doc.Name = "foo"
		err := Save(doc)
		So(err, ShouldEqual, nil)

		Convey("should find one with query", func() {
			newDoc := NewDoc(hookedDocument{}).(*hookedDocument)
			err := newDoc.FindOne(bson.M{
				"name": "foo",
			}, newDoc)
			So(err, ShouldEqual, nil)
			So(newDoc.ID.Hex(), ShouldEqual, doc.ID.Hex())
		})

		Convey("should find one with query and run afterFind", func() {
			newDoc := NewDoc(hookedDocument{}).(*hookedDocument)
			err := newDoc.FindOne(bson.M{
				"name": "foo",
			}, newDoc)
			So(err, ShouldEqual, nil)
			So(newDoc.ID.Hex(), ShouldEqual, doc.ID.Hex())
			So(newDoc.RanAfterFind, ShouldEqual, true)
		})

		Reset(func() {
			conn.Session.DB("mogotest").DropDatabase()
		})
	})

	Convey("Delete", t, func() {
		Convey("should be able delete a document", func() {
			doc := NewDoc(noHookDocument{}).(*noHookDocument)
			doc.Name = "foo"

			err := Save(doc)
			So(err, ShouldEqual, nil)

			err = Remove(doc)
			So(err, ShouldEqual, nil)

			count, err := doc.GetColl().C().Count()

			So(err, ShouldEqual, nil)
			So(count, ShouldEqual, 0)
		})

		Convey("should be able delete a document and run hooks", func() {
			doc := NewDoc(hookedDocument{}).(*hookedDocument)
			doc.Name = "foo"
			doc.Surname = "bar"

			err := Save(doc)
			So(err, ShouldEqual, nil)

			err = Remove(doc)
			So(err, ShouldEqual, nil)

			count, err := doc.GetColl().C().Count()

			So(err, ShouldEqual, nil)
			So(count, ShouldEqual, 0)

			So(doc.RanBeforeDelete, ShouldEqual, true)
			So(doc.RanAfterDelete, ShouldEqual, true)
		})

		Convey("should be able delete a document with RemoveBySelector", func() {
			doc := NewDoc(noHookDocument{}).(*noHookDocument)
			doc.Name = "foo"

			err := Save(doc)
			So(err, ShouldEqual, nil)

			err = RemoveBySelector(doc, bson.M{
				"_id": doc.ID,
			})
			So(err, ShouldEqual, nil)

			count, err := doc.GetColl().C().Count()
			So(err, ShouldEqual, nil)
			So(count, ShouldEqual, 0)
		})

		Convey("should be able delete a document with Remove", func() {
			doc := NewDoc(noHookDocument{}).(*noHookDocument)
			doc.Name = "foo"

			err := Save(doc)
			So(err, ShouldEqual, nil)

			err = doc.Remove()
			So(err, ShouldEqual, nil)

			count, err := doc.GetColl().C().Count()

			So(err, ShouldEqual, nil)
			So(count, ShouldEqual, 0)
		})

	})
}

func TestCollectionWithSlice(t *testing.T) {
	conn := getConnection()
	defer conn.Session.Close()

	Convey("Saving", t, func() {
		Convey("should be able to save a document with no hooks, update id, and use new tracker", func() {
			cMap := make(map[string][]extraData, 0)
			cMap["White"] = []extraData{
				extraData{SubColors: []string{"WhiteWhite", "WhiteYellow"}},
			}
			doc := NewDoc(noHookDocumentWithSlice{
				Name:     "mogo",
				Colors:   []string{"Red", "Green"},
				ColorMap: cMap,
				SubColor: extraData{
					SubColors: []string{"Green"},
				},
				SubColors: []extraData{
					extraData{SubColors: []string{"Pink", "Yellow"}},
					extraData{SubColors: []string{"Black", "White"}},
				},
			}).(*noHookDocumentWithSlice)
			doc.Name = "foo"
			So(doc.IsNew(), ShouldEqual, true)

			err := Save(doc)
			So(err, ShouldEqual, nil)
			So(doc.ID.Valid(), ShouldEqual, true)
			So(doc.IsNew(), ShouldEqual, false)
		})
	})
}
