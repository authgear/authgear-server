package template

import (
	"reflect"
)

// Embed embeds the given value into data.
func Embed(data map[string]any, value any) {
	v := reflect.ValueOf(value)
	typ := v.Type()
	switch typ.Kind() {
	case reflect.Pointer:
		Embed(data, v.Elem().Interface())
	case reflect.Struct:
		numField := typ.NumField()
		for i := range numField {
			structField := typ.Field(i)
			data[structField.Name] = v.Field(i).Interface()
		}
	case reflect.Map:
		if typ.Key().Kind() != reflect.String {
			panic("template: can only embed string-keyed map")
		}
		iter := v.MapRange()
		for iter.Next() {
			k := iter.Key().String()
			v := iter.Value().Interface()
			data[k] = v
		}
	default:
		panic("template: unsupported value kind: " + typ.Kind().String())
	}
}
