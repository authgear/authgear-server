package reflect

import (
	"reflect"
)

// IsRecursivelyZero reports if i is zero recursively.
// nolint: gocyclo
func IsRecursivelyZero(i interface{}) bool {
	v := reflect.ValueOf(i)

	if !v.IsValid() {
		return true
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
	case reflect.Array:
		return v.IsZero()
	case reflect.Chan:
		fallthrough
	case reflect.Func:
		fallthrough
	case reflect.Map:
		fallthrough
	case reflect.Slice:
		return v.IsNil()
	case reflect.Interface:
		fallthrough
	case reflect.Ptr:
		fallthrough
	case reflect.UnsafePointer:
		if v.IsNil() {
			return true
		}
		return IsRecursivelyZero(v.Elem().Interface())
	case reflect.Struct:
		numField := v.NumField()
		for i := 0; i < numField; i++ {
			f := v.Field(i)
			if !IsRecursivelyZero(f.Interface()) {
				return false
			}
		}
		return true
	}

	panic("unreachable")
}
