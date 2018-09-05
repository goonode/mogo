package mogo

import (
	"reflect"

	"github.com/globalsign/mgo/bson"
)

// ValidateRequired ...
func ValidateRequired(val interface{}) bool {
	valueOf := reflect.ValueOf(val)
	return valueOf.Interface() != reflect.Zero(valueOf.Type()).Interface()
}

// ValidateMongoIDRef ...
func ValidateMongoIDRef(id bson.ObjectId, collection *Collection) bool {
	count, err := collection.C().Find(bson.M{"_id": id}).Count()

	if err != nil || count <= 0 {
		return false
	}

	return true
}

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

// ValidateInclusionIn ...
func ValidateInclusionIn(value string, options []string) bool {
	return stringInSlice(value, options)
}
