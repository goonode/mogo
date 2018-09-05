package mogo

import (
	"fmt"
	"reflect"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

type Bongo struct {
	DocumentModel `bson:",inline" coll:"mogo-registry"` // The mogo will be stored in the mogo-registry collection
	Name          string
	Friends       RefFieldSlice `ref:"Macao"` // The field Friends of mogo is a reference to a slice of Macao objects
	BestFriend    RefField      `ref:"Macao"`
}

type Macao struct {
	DocumentModel `bson:",inline" coll:"mogo-registry"` // The Macao will be stored in the mogo-registry collection
	Name          string
}

func TestRegister(t *testing.T) {
	var mr ModelReg

	mr.Register(noHookDocument{},
		hookedDocument{})
	Convey("should register the passed interfaces", t, func() {
		n, _, b := mr.Exists(noHookDocument{})
		So(n, ShouldEqual, "noHookDocument")
		So(b, ShouldBeTrue)
		n, _, b = mr.Exists(hookedDocument{})
		So(n, ShouldEqual, "hookedDocument")
		So(b, ShouldBeTrue)
		n, _, b = mr.Exists(DocumentChild{})
		So(n, ShouldEqual, "")
		So(b, ShouldBeFalse)
	})

	Convey("should not register a struct that not has DocumentModel", t, func() {
		So(func() { mr.Register(BadDocument{}) }, ShouldPanic)
	})

	Convey("should return the index of the DocumentModel field and the Type", t, func() {
		i := mr.Index("hookedDocument")
		t := mr.TypeOf("hookedDocument")
		So(i, ShouldEqual, 0)
		So(t, ShouldResemble, reflect.TypeOf(hookedDocument{}))
	})

}

func TestRegisterRef(t *testing.T) {
	Convey("should register the passed interfaces", t, func() {
		ModelRegistry.Register(Bongo{}, Macao{})
		_, _, ok := ModelRegistry.Exists(Bongo{})
		So(ok, ShouldBeTrue)
		_, _, ok = ModelRegistry.Exists(Macao{})
		So(ok, ShouldBeTrue)
		So(ModelRegistry["mogo"].Refs["Friends"].Ref, ShouldEqual, "Macao")
		So(ModelRegistry["mogo"].Refs["Friends"].Exists, ShouldBeTrue)
	})
}

func TestNewFunc(t *testing.T) {
	ModelRegistry.Register(Bongo{}, Macao{})

	macao := ModelRegistry.New("Macao").(*Macao)
	fmt.Println(macao.Name)
}

func TestInterfaceNameFunc(t *testing.T) {
	var mr ModelReg

	mr.Register(noHookDocument{},
		hookedDocument{})
	Convey("should return the name of the passed name", t, func() {
		n := interfaceName(map[string]*noHookDocument{})
		So(n, ShouldEqual, "noHookDocument")
		n = interfaceName(map[string]noHookDocument{})
		So(n, ShouldEqual, "noHookDocument")
		n = interfaceName(&[]*noHookDocument{})
		So(n, ShouldEqual, "noHookDocument")
		n = interfaceName([]*noHookDocument{})
		So(n, ShouldEqual, "noHookDocument")
		n = interfaceName([]noHookDocument{})
		So(n, ShouldEqual, "noHookDocument")
		n = interfaceName(&noHookDocument{})
		So(n, ShouldEqual, "noHookDocument")
		n = interfaceName(noHookDocument{})
		So(n, ShouldEqual, "noHookDocument")
	})

}
