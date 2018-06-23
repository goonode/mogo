package bongo

import (
	"fmt"
	"reflect"
)

// Registry ...
type Registry interface {
	Register(...interface{})
	Exists(interface{}) (string, *ModelInternals, bool)

	Index(string) int
	TypeOf(string) reflect.Type

	// New(interface{}) interface{}
}

// ModelInternals contains some internal information about the model
type ModelInternals struct {
	// Idx is the index of the field containing the DM
	Idx int
	// The Type
	Type reflect.Type
}

// ModelReg ...
type ModelReg map[string]*ModelInternals

// modelRegistry is the centralized registry of all models used for the app
var modelRegistry = make(ModelReg, 0)

// Register ...
func (r ModelReg) Register(i ...interface{}) {
	for p, o := range i {
		t := reflect.TypeOf(o)
		v := reflect.ValueOf(o)
		n := t.Name()

		if t.Kind() == reflect.Ptr {
			t = reflect.Indirect(reflect.ValueOf(o)).Type()
			v = reflect.ValueOf(0).Elem()
		}
		if t.Kind() != reflect.Struct {
			panic(fmt.Sprintf("Only type struct can be used as document model (passed type %s (pos: %d) is not struct)", n, p))
		}
		var Idx = -1
		for i := 0; i < v.NumField(); i++ {
			ft := t.Field(i)
			if ft.Type.ConvertibleTo(reflect.TypeOf(DocumentModel{})) {
				Idx = i
				break
			}
		}

		if Idx == -1 {
			panic(fmt.Sprintf("A document model must embed a DocumentModel type field (passed type %s (pos: %d) does not have)", n, p))
		}

		modelRegistry[n] = &ModelInternals{Idx: Idx, Type: t}
	}
}

// Exists ...
func (r ModelReg) Exists(i interface{}) (string, *ModelInternals, bool) {
	n := reflect.TypeOf(i).Name()
	if t, ok := modelRegistry[n]; ok {
		return n, t, true
	}
	return "", nil, false
}

// TypeOf ...
func (r ModelReg) TypeOf(n string) reflect.Type {
	if v, ok := modelRegistry[n]; ok {
		return v.Type
	}
	return nil
}

// Index returns the index of the DocumentModel field in the struct
// or -1 if the struct name passed is not found
func (r ModelReg) Index(n string) int {
	if v, ok := modelRegistry[n]; ok {
		return v.Idx
	}
	return -1
}

// New ...
// func (r ModelReg) New(i interface{}) interface{} {
// 	n := reflect.TypeOf(i).Name()
// 	if t, ok := modelRegistry[n]; ok {
// 		return reflect.New(t.Type).Elem().Interface()
// 	}

// 	return nil
// }
