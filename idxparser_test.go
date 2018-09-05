package mogo

import (
	"testing"

	"github.com/globalsign/mgo"
	. "github.com/smartystreets/goconvey/convey"
)

func TestScan(t *testing.T) {
	Convey("should return the tokens in passed []byte", t, func() {
		r := []ParsedIndex{
			ParsedIndex{[]string{"name", "surname"}, []string{"unique", "sparse"}, 0, false},
			ParsedIndex{[]string{"surname"}, []string{"unique"}, 0, false},
		}
		p := IndexScan("{name,surname},unique,sparse;{surname},unique")
		So(p, ShouldResemble, r)
		So(func() { IndexScan("{},unique,sparse;{surname},unique") }, ShouldPanic)
	})
}

func TestBuildIndex(t *testing.T) {
	Convey("should return the Index struct for each ParsedIndex", t, func() {
		p := IndexScan("{name,surname},unique,sparse;{surname},unique")
		var idxes []*mgo.Index
		for i := range p {
			idxes = append(idxes, BuildIndex(p[i]))
		}
	})
}
