package bongo

import (
	"reflect"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

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

	// Convey("should make a new instance of the passed model", t, func() {
	// 	h := mr.New(hookedDocument{}).(*hookedDocument)
	// 	So(reflect.TypeOf(h), ShouldResemble, reflect.TypeOf(hookedDocument{}))
	// })

}
