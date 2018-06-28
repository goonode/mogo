package bongo

import (
	"reflect"
	"testing"
	"time"

	"github.com/globalsign/mgo/bson"
	. "github.com/smartystreets/goconvey/convey"
)

type CascadeParent struct {
	DocumentModel `bson:",inline" coll:"parents"`
	Bar           string
	Number        int
	FooBar        string
	Children      []ChildRef
	Child         ChildRef
	ChildProp     string `bson:"childProp"`
	diffTracker   *DiffTracker
}

func (f *CascadeParent) GetDiffTracker() *DiffTracker {
	v := reflect.ValueOf(f.diffTracker)
	if !v.IsValid() || v.IsNil() {
		f.diffTracker = NewDiffTracker(f)
	}

	return f.diffTracker
}

type CascadeChild struct {
	DocumentModel `bson:",inline" coll:"children"`
	ParentID      bson.ObjectId `bson:",omitempty"`
	Name          string
	SubChild      SubChildRef `bson:"subChild"`
	ChildProp     string
	diffTracker   *DiffTracker
}

func (c *CascadeChild) GetCascade(collection *Collection) []*CascadeConfig {
	ref := ChildRef{
		ID:       c.ID,
		Name:     c.Name,
		SubChild: c.SubChild,
	}

	cascadeSingle := &CascadeConfig{
		Collection:  c.GetColl(),
		Properties:  []string{"_id", "name", "subChild.foo", "subChild._id"},
		Data:        ref,
		ThroughProp: "child",
		RelType:     RelOne,
		Query: bson.M{
			"_id": c.ParentID,
		},
	}

	cascadeCopy := &CascadeConfig{
		Collection: c.GetColl(),
		Properties: []string{"childProp"},
		Data: map[string]interface{}{
			"childProp": c.ChildProp,
		},
		RelType: RelOne,
		Query: bson.M{
			"_id": c.ParentID,
		},
	}

	cascadeMulti := &CascadeConfig{
		Collection:  c.GetColl(),
		Properties:  []string{"_id", "name", "subChild.foo", "subChild._id"},
		Data:        ref,
		ThroughProp: "children",
		RelType:     RelMany,
		Query: bson.M{
			"_id": c.ParentID,
		},
	}

	if c.GetDiffTracker().Modified("ParentID") {
		origID, _ := c.diffTracker.GetOriginalValue("ParentID")
		if origID != nil {
			oldQuery := bson.M{
				"_id": origID,
			}
			cascadeSingle.OldQuery = oldQuery
			cascadeCopy.OldQuery = oldQuery
			cascadeMulti.OldQuery = oldQuery
		}

	}

	return []*CascadeConfig{cascadeSingle, cascadeMulti, cascadeCopy}
}

func (c *CascadeChild) GetDiffTracker() *DiffTracker {
	v := reflect.ValueOf(c.diffTracker)
	if !v.IsValid() || v.IsNil() {
		c.diffTracker = NewDiffTracker(c)
	}

	return c.diffTracker
}

type SubChild struct {
	DocumentModel `bson:",inline" coll:"subchildren"`
	Foo           string
	ChildID       bson.ObjectId
}

func (c *SubChild) GetCascade(collection *Collection) []*CascadeConfig {
	ref := SubChildRef{
		ID:  c.ID,
		Foo: c.Foo,
	}
	cascadeSingle := &CascadeConfig{
		Collection:  c.GetColl(),
		Properties:  []string{"_id", "foo"},
		Data:        ref,
		ThroughProp: "subChild",
		RelType:     RelOne,
		Query: bson.M{
			"_id": c.ChildID,
		},
		Nest:     true,
		Instance: &CascadeChild{},
	}

	return []*CascadeConfig{cascadeSingle}
}

type SubChildRef struct {
	ID  bson.ObjectId `bson:"_id,omitempty"`
	Foo string
}

type ChildRef struct {
	ID       bson.ObjectId `bson:"_id,omitempty"`
	Name     string
	SubChild SubChildRef
}

func TestCascade(t *testing.T) {
	connection := getConnection()
	// defer connection.Session.Close()

	Convey("Cascade Save/Delete - full runthrough", t, func() {
		connection.Session.DB("bongotest").DropDatabase()
		parent := NewDocument(CascadeParent{}).(*CascadeParent)
		parent.Bar = "Testy McGee"
		parent.Number = 5

		parent2 := NewDocument(CascadeParent{}).(*CascadeParent)
		parent2.Bar = "Other Parent"
		parent2.Number = 10

		err := Save(parent)
		So(err, ShouldEqual, nil)
		err = Save(parent2)
		So(err, ShouldEqual, nil)

		child := NewDocument(CascadeChild{}).(*CascadeChild)
		child.ParentID = parent.ID
		child.Name = "Foo McGoo"
		child.ChildProp = "Doop McGoop"
		err = Save(child)

		// Wait a sec for the go routine to finish.
		time.Sleep(100 * time.Millisecond)

		So(err, ShouldEqual, nil)

		child.GetDiffTracker().Reset()
		newParent := NewDocument(CascadeParent{}).(*CascadeParent)
		FindByID(newParent, parent.ID)

		So(newParent.Child.Name, ShouldEqual, "Foo McGoo")
		So(newParent.Child.ID.Hex(), ShouldEqual, child.ID.Hex())
		So(newParent.Children[0].Name, ShouldEqual, "Foo McGoo")
		So(newParent.Children[0].ID.Hex(), ShouldEqual, child.ID.Hex())

		// No through prop should populate directly o the parent
		So(newParent.ChildProp, ShouldEqual, "Doop McGoop")

		// Now change the child parent Id...
		child.ParentID = parent2.ID
		So(child.GetDiffTracker().Modified("ParentID"), ShouldEqual, true)

		err = Save(child)
		So(err, ShouldEqual, nil)

		// Wait a sec for the go routine to finish.
		time.Sleep(100 * time.Millisecond)

		child.diffTracker.Reset()
		// Now make sure it says the parent id DIDNT change, because we just reset the tracker
		So(child.GetDiffTracker().Modified("ParentID"), ShouldEqual, false)

		newParent1 := NewDocument(CascadeParent{}).(*CascadeParent)
		FindByID(newParent1, parent.ID)
		So(newParent1.Child.Name, ShouldEqual, "")
		So(newParent1.ChildProp, ShouldEqual, "")
		So(len(newParent1.Children), ShouldEqual, 0)
		newParent2 := NewDocument(CascadeParent{}).(*CascadeParent)
		FindByID(newParent2, parent2.ID)
		So(newParent2.ChildProp, ShouldEqual, "Doop McGoop")
		So(newParent2.Child.Name, ShouldEqual, "Foo McGoo")
		So(newParent2.Child.ID.Hex(), ShouldEqual, child.ID.Hex())
		So(newParent2.Children[0].Name, ShouldEqual, "Foo McGoo")
		So(newParent2.Children[0].ID.Hex(), ShouldEqual, child.ID.Hex())

		// Make a new sub child, save it, and it should cascade to the child AND the parent
		subChild := NewDocument(SubChild{}).(*SubChild)
		subChild.Foo = "MySubChild"
		subChild.ChildID = child.ID
		err = Save(subChild)
		So(err, ShouldEqual, nil)

		// Wait a sec for the go routine to finish.
		time.Sleep(100 * time.Millisecond)

		// Fetch the parent
		newParent3 := NewDocument(CascadeParent{}).(*CascadeParent)
		FindByID(newParent3, parent2.ID)
		So(newParent3.Child.SubChild.Foo, ShouldEqual, "MySubChild")
		So(newParent3.Child.SubChild.ID.Hex(), ShouldEqual, subChild.ID.Hex())

		newParent4 := NewDocument(CascadeParent{}).(*CascadeParent)
		err = DeleteDocument(child)

		// Wait a sec for the go routine to finish.
		time.Sleep(100 * time.Millisecond)

		So(err, ShouldEqual, nil)
		FindByID(newParent4, parent2.ID)
		So(newParent4.Child.Name, ShouldEqual, "")
		So(newParent4.ChildProp, ShouldEqual, "")
		So(len(newParent4.Children), ShouldEqual, 0)
	})

	Convey("MapFromCascadeProperties", t, func() {
		parent := NewDocument(CascadeParent{}).(*CascadeParent)
		parent.Bar = "bar"
		parent.Child = ChildRef{
			Name: "child",
			SubChild: SubChildRef{
				Foo: "foo",
			},
		}
		parent.Number = 5
		props := []string{"bar", "child.name"}

		mp := MapFromCascadeProperties(props, parent)

		So(len(mp), ShouldEqual, 2)
		So(mp["bar"], ShouldEqual, "bar")

		submp := mp["child"].(map[string]interface{})
		So(submp["name"], ShouldEqual, "child")
	})
}
