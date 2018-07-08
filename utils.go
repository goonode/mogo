package bongo

import (
	"reflect"
	"strings"
	"unicode"
)

//Lower cases first char of string
func lowerInitial(str string) string {
	for i, v := range str {
		return string(unicode.ToLower(v)) + str[i+1:]
	}
	return ""
}

// GetBsonName ...
func GetBsonName(field reflect.StructField) string {
	tag := field.Tag.Get("bson")
	tags := strings.Split(tag, ",")

	if len(tags[0]) > 0 {
		return tags[0]
	}

	return lowerInitial(field.Name)
}

// ValueOf return the reflect Value of d
func ValueOf(d interface{}) reflect.Value {
	var v reflect.Value

	if reflect.TypeOf(d).Kind() == reflect.Ptr {
		v = reflect.ValueOf(d).Elem()
	} else {
		v = reflect.ValueOf(d)
	}

	return v
}
