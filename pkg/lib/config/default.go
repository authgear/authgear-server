package config

import (
	"reflect"
)

type defaulter interface {
	SetDefaults()
}

func set(t reflect.Type, v reflect.Value) {
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

func setFieldDefaults(value interface{}) {
	t := reflect.TypeOf(value)
	v := reflect.ValueOf(value)
	set(t, v)
}

func unsetFieldDefaults(value interface{}, defaults interface{}) {
	t := reflect.TypeOf(value)
	v1 := reflect.ValueOf(value)
	v2 := reflect.ValueOf(defaults)

	if t.Kind() == reflect.Ptr && t.Elem().Kind() == reflect.Ptr && t.Elem().Elem().Kind() == reflect.Struct {
		v1Elem := v1.Elem().Elem()
		v2Elem := v2.Elem().Elem()
		numField := t.Elem().Elem().NumField()
		if v1Elem.IsValid() && v2Elem.IsValid() {
			for j := 0; j < numField; j++ {
				v1Field := v1Elem.Field(j)
				v2Field := v2Elem.Field(j)
				unsetFieldDefaults(v1Field.Addr().Interface(), v2Field.Addr().Interface())
			}
		}

		// If the struct has SetDefaults() and the value is equal to defaults,
		// reset the struct to a nil pointer.
		_, ok := v1.Elem().Interface().(defaulter)
		if ok {
			eq := reflect.DeepEqual(value, defaults)
			if eq {
				if v1.Elem().CanSet() {
					v1.Elem().Set(reflect.Zero(t.Elem()))
					v2.Elem().Set(reflect.Zero(t.Elem()))
				}
			}
		}

		// If the struct does not support SetDefaults() and the value is zero,
		// reset the struct to a nil pointer.
		if !ok && v1.Elem().Elem().IsZero() {
			if v1.Elem().CanSet() {
				v1.Elem().Set(reflect.Zero(t.Elem()))
				v2.Elem().Set(reflect.Zero(t.Elem()))
			}
		}
	}
}
