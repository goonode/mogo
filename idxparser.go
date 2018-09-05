package mogo

import (
	"go/scanner"
	"go/token"
	"reflect"
	"strings"

	"github.com/globalsign/mgo"
)

var optionKeywords = [...]string{"unique", "sparse", "background", "dropdups"}

// ParsedIndex contains a parsed index
type ParsedIndex struct {
	Fields  []string
	Options []string

	lastFieldIdx int
	stickyField  bool
}

// RefIndex contains the object stored as reference in database
type RefIndex struct {
	// The docuemnt model name
	Model string

	// The referenced object name
	Ref string

	// The field index in the parsed struct
	Idx int

	// The kind of the field (slice or other)
	Kind reflect.Kind

	// The type of the field
	Type reflect.Type

	// Whenever the reference object exists in the Registry
	Exists bool
}

// IndexScan ...
// TODO:
// 	add optional parameter to pass the name of the field to be used
// 	in case of empty {} or if filled name is empty
func IndexScan(src string) []ParsedIndex {
	var s scanner.Scanner
	var parsed []ParsedIndex

	src = TrimAllSpaces(src)
	fset := token.NewFileSet()
	file := fset.AddFile("", fset.Base(), len(src))
	s.Init(file, []byte(src), nil /* no error handler */, scanner.ScanComments)

	// Repeated calls to Scan yield the token sequence found in the input.
	lb := false
	p := &ParsedIndex{}
	for {
		_, tok, lit := s.Scan()

		switch tok {
		case token.LBRACE:
			if lb {
				goto _panic
			}
			lb = true
		case token.RBRACE:
			if !lb || len(p.Fields) == 0 {
				goto _panic
			}
			lb = false
		case token.IDENT:
			if lb {
				if p.getStickyField() {
					p.appendDotField(lit)
					p.setStickyFieldTo(false)
					break
				}

				p.appendField(lit)
				break
			}
			p.appendOption(lit)
		case token.PERIOD:
			if p.getStickyField() {
				goto _panic
			}
			p.setStickyFieldTo(true)
			p.updateLastFieldIdx()
		case token.COMMA:
		case token.COLON:
			if lb {
				goto _panic
			}
		case token.SEMICOLON:
			if lb {
				goto _panic
			}
			parsed = append(parsed, *p)
			p = &ParsedIndex{}
		case token.EOF:
			if lb {
				goto _panic
			}
			return parsed
		default:
			goto _panic
		}
	}

_panic:
	panic("Syntax error in parsing index expression")
}

// BuildIndex build an mgo Index using the values of a ParsedIndex
// struct
func BuildIndex(p ParsedIndex) *mgo.Index {
	idx := &mgo.Index{
		Key: p.Fields,
	}

	for i := range p.Options {
		switch p.Options[i] {
		case "unique":
			idx.Unique = true
		case "dropdups":
			idx.DropDups = true
		case "background":
			idx.Background = true
		case "sparse":
			idx.Sparse = true
		}
	}

	return idx
}

func (p *ParsedIndex) appendOption(o string) {
	o = strings.ToLower(o)
	o = strings.Trim(o, " ")
	for i := range optionKeywords {
		if optionKeywords[i] == o {
			p.Options = append(p.Options, o)
		}
	}
}

func (p *ParsedIndex) appendField(f string) {
	p.Fields = append(p.Fields, f)
}
func (p *ParsedIndex) appendDotField(f string) {
	p.Fields[p.lastFieldIdx] = p.Fields[p.lastFieldIdx] + "." + f
}
func (p *ParsedIndex) updateLastFieldIdx() {
	if len(p.Fields) > 0 {
		p.lastFieldIdx = len(p.Fields) - 1
		return
	}

	p.lastFieldIdx = 0
}
func (p *ParsedIndex) setStickyFieldTo(s bool) {
	p.stickyField = s
}

func (p *ParsedIndex) getStickyField() bool {
	return p.stickyField
}
