package bongo

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestScan(t *testing.T) {
	Convey("should return the tokens in passed []byte", t, func() {
		r := []ParsedIndex{
			ParsedIndex{[]string{"name", "surname"}, []string{"unique", "sparse"}},
			ParsedIndex{[]string{"surname"}, []string{"unique"}},
		}
		p := Scan("{name,surname},unique,sparse;{surname},unique")
		So(p, ShouldResemble, r)
		So(func() { Scan("{},unique,sparse;{surname},unique") }, ShouldPanic)
	})
}
