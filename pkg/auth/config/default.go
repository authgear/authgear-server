package config

import "reflect"

type defaulter interface {
	SetDefaults()
}

func setFieldDefaults(value interface{}) {
	var set func(t reflect.Type, v reflect.Value)
	set = func(t reflect.Type, v reflect.Value) {
		switch t.Kind() {
		case reflect.Slice:
			itemT := reflect.PtrTo(t.Elem())
			for i := 0; i < v.Len(); i++ {
				item := v.Index(i).Addr()
				set(itemT, item)
			}
		case reflect.Struct:
			numField := t.NumField()
			for j := 0; j < numField; j++ {
				field := v.Field(j)
				ft := t.Field(j)
				set(ft.Type, field)
			}
		case reflect.Ptr:
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
