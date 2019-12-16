package marshal

import (
	"reflect"

	coreReflect "github.com/skygeario/skygear-server/pkg/core/reflect"
)

// UpdateNilFieldsWithZeroValue checks the fields with tag
// `default_zero_value:"true"` and updates the fields with zero value if they are nil.
// This function will walk through the struct recursively,
// if the tagged fields of struct have duplicated type in the same path,
// the function may cause infinite loop.
// Before calling this function, please make sure the struct get pass with
// function `shouldNotHaveDuplicatedTypeInSamePath` in the test case.
func UpdateNilFieldsWithZeroValue(i interface{}) {
	t := reflect.TypeOf(i).Elem()
	v := reflect.ValueOf(i).Elem()

	if t.Kind() != reflect.Struct {
		return
	}
	numField := t.NumField()
	for i := 0; i < numField; i++ {
		zerovalueTag := t.Field(i).Tag.Get("default_zero_value")
		if zerovalueTag != "true" {
			continue
		}

		field := v.Field(i)
		ft := t.Field(i)
		if field.Kind() == reflect.Ptr {
			ele := field.Elem()
			if !ele.IsValid() {
				ele = reflect.New(ft.Type.Elem())
				field.Set(ele)
			}
			UpdateNilFieldsWithZeroValue(field.Interface())
		}
	}
}

// OmitEmpty traverses the given struct and set pointer to struct to nil if
// the value IsRecursivelyZero.
func OmitEmpty(i interface{}) {
	OmitEmptyValue(reflect.ValueOf(i))
}

// nolint: gocyclo
func OmitEmptyValue(v reflect.Value) {
	if !v.IsValid() {
		return
	}

	switch v.Type().Kind() {
	case reflect.Bool:
		fallthrough
	case reflect.Int:
		fallthrough
	case reflect.Int8:
		fallthrough
	case reflect.Int16:
		fallthrough
	case reflect.Int32:
		fallthrough
	case reflect.Int64:
		fallthrough
	case reflect.Uint:
		fallthrough
	case reflect.Uint8:
		fallthrough
	case reflect.Uint16:
		fallthrough
	case reflect.Uint32:
		fallthrough
	case reflect.Uint64:
		fallthrough
	case reflect.Uintptr:
		fallthrough
	case reflect.Float32:
		fallthrough
	case reflect.Float64:
		fallthrough
	case reflect.Complex64:
		fallthrough
	case reflect.Complex128:
		fallthrough
	case reflect.String:
		fallthrough
	case reflect.Chan:
		fallthrough
	case reflect.Func:
		// Primitives can be omit-empty automatically.
		return
	case reflect.Array:
		numItems := v.Len()
		for i := 0; i < numItems; i++ {
			elem := v.Index(i)
			OmitEmptyValue(elem)
		}
	case reflect.Map:
		iter := v.MapRange()
		for iter.Next() {
			if coreReflect.IsRecursivelyZero(iter.Value().Interface()) {
				v.SetMapIndex(iter.Key(), reflect.Value{})
			} else {
				OmitEmptyValue(iter.Value())
			}
		}
	case reflect.Slice:
		numItems := v.Len()
		for i := 0; i < numItems; i++ {
			elem := v.Index(i)
			if coreReflect.IsRecursivelyZero(elem.Interface()) {
				elem.Set(reflect.Zero(elem.Type()))
			} else {
				OmitEmptyValue(elem)
			}
		}
	case reflect.Interface:
		fallthrough
	case reflect.Ptr:
		fallthrough
	case reflect.UnsafePointer:
		if v.IsNil() {
			return
		}
		OmitEmptyValue(v.Elem())
	case reflect.Struct:
		numField := v.NumField()
		for i := 0; i < numField; i++ {
			f := v.Field(i)
			if coreReflect.IsRecursivelyZero(f.Interface()) {
				f.Set(reflect.Zero(f.Type()))
			} else {
				OmitEmptyValue(f)
			}
		}
	}
}

func ShouldNotHaveDuplicatedTypeInSamePath(i interface{}, pathSet map[string]interface{}) bool {
	t := reflect.TypeOf(i).Elem()
	v := reflect.ValueOf(i).Elem()

	if t.Kind() != reflect.Struct {
		return true
	}
	numField := t.NumField()
	for i := 0; i < numField; i++ {
		zerovalueTag := t.Field(i).Tag.Get("default_zero_value")
		if zerovalueTag != "true" {
			continue
		}

		field := v.Field(i)
		ft := t.Field(i)
		if field.Kind() == reflect.Ptr {
			ele := field.Elem()
			if !ele.IsValid() {
				ele = reflect.New(ft.Type.Elem())
				field.Set(ele)
			}
			typeName := ft.Type.String()
			if _, ok := pathSet[typeName]; ok {
				return false
			}
			newSet := copySet(pathSet)
			newSet[ft.Type.String()] = struct{}{}
			pass := ShouldNotHaveDuplicatedTypeInSamePath(field.Interface(), newSet)
			if !pass {
				return false
			}
		}
	}

	return true
}

func copySet(input map[string]interface{}) map[string]interface{} {
	output := map[string]interface{}{}
	for k := range input {
		output[k] = input[k]
	}

	return output
}
