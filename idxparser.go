package bongo

import (
	"go/scanner"
	"go/token"
	"strings"
	"unicode"

	"github.com/globalsign/mgo"
)

var optionKeywords = [...]string{"unique", "sparse", "background", "dropdups"}

// ParsedIndex contains a parsed index
type ParsedIndex struct {
	Fields  []string
	Options []string
}

// TrimAllSpaces removes all spaces from the passed string and
// returns the trimmed string
func TrimAllSpaces(src string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsSpace(r) {
			return -1
		}
		return r
	}, src)
}

// IndexScan ...
// TODO:
// 	add optional parameter to pass the name of the field to be used
// 	in case of empty {} or if filled name is empty
func IndexScan(src string) []ParsedIndex {
	var s scanner.Scanner
	var parsed []ParsedIndex

	fset := token.NewFileSet()
	file := fset.AddFile("", fset.Base(), len(src))
	s.Init(file, []byte(TrimAllSpaces(src)), nil /* no error handler */, scanner.ScanComments)

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
				p.appendField(lit)
				break
			}
			p.appendOption(lit)
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
