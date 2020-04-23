package marshal

import (
	"reflect"
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
	updateNilFieldsWithZeroValue(t, v, reflect.StructTag(""))
}

func updateNilFieldsWithZeroValue(t reflect.Type, v reflect.Value, st reflect.StructTag) {
	shouldSetZeroValue := st.Get("default_zero_value") == "true"

	switch v.Kind() {
	case reflect.Slice:
		if shouldSetZeroValue && v.IsNil() {
			v.Set(reflect.MakeSlice(t, 0, 0))
		}
		subt := t.Elem()
		for i := 0; i < v.Len(); i++ {
			item := v.Index(i)
			updateNilFieldsWithZeroValue(subt, item, reflect.StructTag(""))
		}
	case reflect.Struct:
		numField := t.NumField()
		for j := 0; j < numField; j++ {
			field := v.Field(j)
			ft := t.Field(j)
			updateNilFieldsWithZeroValue(ft.Type, field, ft.Tag)
		}
	case reflect.Ptr:
		ele := v.Elem()
		if shouldSetZeroValue {
			if !ele.IsValid() {
				ele = reflect.New(t.Elem())
				v.Set(ele)
			}
		}
		if ele.IsValid() {
			i := v.Interface()
			t := reflect.TypeOf(i).Elem()
			v := reflect.ValueOf(i).Elem()
			updateNilFieldsWithZeroValue(t, v, reflect.StructTag(""))
		}
	case reflect.Map:
		if shouldSetZeroValue && v.IsNil() {
			v.Set(reflect.MakeMap(t))
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
