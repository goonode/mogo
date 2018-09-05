package mogo

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

// ValueOf return the reflect Value of d. In case of slice or map
// it reduces to a new primitive type.
func ValueOf(d interface{}) reflect.Value {
	v := reflect.ValueOf(d)

	if v.Type().Kind() == reflect.Slice || v.Type().Kind() == reflect.Map {
		inner := v.Type().Elem()
		switch inner.Kind() {
		case reflect.Ptr:
			v = reflect.New(inner.Elem()).Elem()
		default:
			v = reflect.New(inner).Elem()
		}
	} else if v.Type().Kind() == reflect.Ptr {
		return ValueOf(reflect.Indirect(v).Interface())
	}

	return v
}

func isSlice(s interface{}) bool {
	if reflect.TypeOf(s).Kind() != reflect.Slice {
		return false
	}

	return true
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
