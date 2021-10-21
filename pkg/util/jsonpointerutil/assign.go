package jsonpointerutil

import (
	"fmt"
	"reflect"

	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"
)

// AssignToJSONObject treats all references tokens of t are map keys.
// It assigns value at t in target.
func AssignToJSONObject(t jsonpointer.T, target interface{}, value interface{}) (err error) {
	if len(t) == 0 {
		err = fmt.Errorf("assigning to root is not supported")
		return
	}

	reflectValue := reflect.ValueOf(target)
	head := t[0]

	if len(t) == 1 {
		switch reflectValue.Kind() {
		case reflect.Map:
			reflectValue.SetMapIndex(reflect.ValueOf(head), reflect.ValueOf(value))
			return
		default:
			err = fmt.Errorf("%v (%T) is not supported", target, target)
			return
		}
	}

	switch reflectValue.Kind() {
	case reflect.Map:
		child := reflectValue.MapIndex(reflect.ValueOf(head))
		if !child.IsValid() {
			sampleMap := map[string]interface{}{}
			m := reflect.MakeMap(reflect.TypeOf(sampleMap))
			reflectValue.SetMapIndex(reflect.ValueOf(head), m)
		}
	default:
		err = fmt.Errorf("%v (%T) is not supported", target, target)
		return
	}

	// recur
	tail := t[1:]
	err = AssignToJSONObject(tail, reflectValue.MapIndex(reflect.ValueOf(head)).Interface(), value)
	return
}

func RemoveFromJSONObject(t jsonpointer.T, target interface{}) (err error) {
	if len(t) == 0 {
		err = fmt.Errorf("removing from root is not supported")
		return
	}

	reflectValue := reflect.ValueOf(target)
	head := t[0]

	if len(t) == 1 {
		switch reflectValue.Kind() {
		case reflect.Map:
			// SetMapIndex says setting a zero Value to delete the map key.
			var zeroValue reflect.Value
			reflectValue.SetMapIndex(reflect.ValueOf(head), zeroValue)
			return
		default:
			err = fmt.Errorf("%v (%T) is not supported", target, target)
			return
		}
	}

	switch reflectValue.Kind() {
	case reflect.Map:
		child := reflectValue.MapIndex(reflect.ValueOf(head))
		if !child.IsValid() {
			sampleMap := map[string]interface{}{}
			m := reflect.MakeMap(reflect.TypeOf(sampleMap))
			reflectValue.SetMapIndex(reflect.ValueOf(head), m)
		}
	default:
		err = fmt.Errorf("%v (%T) is not supported", target, target)
		return
	}

	// recur
	tail := t[1:]
	err = RemoveFromJSONObject(tail, reflectValue.MapIndex(reflect.ValueOf(head)).Interface())
	return
}
