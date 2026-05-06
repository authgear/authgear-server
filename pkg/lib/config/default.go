package config

import (
	"reflect"
)

type defaulter interface {
	SetDefaults()
}

func SetFieldDefaults(value any) {
	var set func(t reflect.Type, v reflect.Value)

	set = func(t reflect.Type, v reflect.Value) {
		switch t.Kind() {
		case reflect.Slice:
			itemT := reflect.PointerTo(t.Elem())
			for i := 0; i < v.Len(); i++ {
				item := v.Index(i).Addr()
				set(itemT, item)
			}
		case reflect.Struct:
			numField := t.NumField()
			for j := range numField {
				field := v.Field(j)
				ft := t.Field(j)
				isNullable := ft.Tag.Get("nullable") == "true"
				// NOTE: builtin structs e.g. time.Time have non-exported fields that cannot be set
				if !isNullable && field.CanSet() {
					set(ft.Type, field)
				}
			}
		case reflect.Pointer:
			ele := v.Elem()
			if !ele.IsValid() && t.Elem().Kind() == reflect.Struct {
				ele = reflect.New(t.Elem())
				v.Set(ele)
			}

			if ele.IsValid() {
				set(ele.Type(), ele)
			}

			i := v.Interface()
			if d, ok := i.(defaulter); ok {
				d.SetDefaults()
			}
		case reflect.Map:
			if v.IsNil() {
				v.Set(reflect.MakeMap(t))
			}

			i := v.Interface()
			if d, ok := i.(defaulter); ok {
				d.SetDefaults()
			}
		}
	}

	t := reflect.TypeOf(value)
	v := reflect.ValueOf(value)
	set(t, v)
}
