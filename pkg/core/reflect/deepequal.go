package reflect

import (
	goreflect "reflect"
)

// NonRecursiveDataDeepEqual compares Go values.
// The differences from reflect.DeepEqual are
//
// - This function does not support recursive data structure,
//   in that case this function will hang.
//
// - It treats nil slice and empty slice equal.
//
// - It treats nil map and empty map equal.
func NonRecursiveDataDeepEqual(a interface{}, b interface{}) bool {
	v1 := goreflect.ValueOf(a)
	v2 := goreflect.ValueOf(b)
	return recur(v1, v2)
}

func recur(v1, v2 goreflect.Value) bool {
	// either one is nil
	if !v1.IsValid() || !v2.IsValid() {
		switch {
		case !v1.IsValid() && !v2.IsValid():
			return true
		case !v1.IsValid() && v2.Kind() == goreflect.Slice && v2.Len() == 0:
			return true
		case v1.Kind() == goreflect.Slice && v1.Len() == 0 && !v2.IsValid():
			return true
		case !v1.IsValid() && v2.Kind() == goreflect.Map && v2.Len() == 0:
			return true
		case v1.Kind() == goreflect.Map && v1.Len() == 0 && !v2.IsValid():
			return true
		default:
			return false
		}
	}

	if v1.Type() != v2.Type() {
		return false
	}

	switch v1.Kind() {
	case goreflect.Array:
		for i := 0; i < v1.Len(); i++ {
			if !recur(v1.Index(i), v2.Index(i)) {
				return false
			}
		}
		return true
	case goreflect.Slice:
		if v1.Pointer() == v2.Pointer() {
			return true
		}
		if v1.Len() != v2.Len() {
			return false
		}
		for i := 0; i < v1.Len(); i++ {
			if !recur(v1.Index(i), v2.Index(i)) {
				return false
			}
		}
		return true
	case goreflect.Interface:
		if v1.IsNil() || v2.IsNil() {
			return v1.IsNil() == v2.IsNil()
		}
		return recur(v1.Elem(), v2.Elem())
	case goreflect.Ptr:
		if v1.Pointer() == v2.Pointer() {
			return true
		}
		return recur(v1.Elem(), v2.Elem())
	case goreflect.Struct:
		for i, n := 0, v1.NumField(); i < n; i++ {
			if !recur(v1.Field(i), v2.Field(i)) {
				return false
			}
		}
		return true
	case goreflect.Map:
		if v1.Pointer() == v2.Pointer() {
			return true
		}
		if v1.Len() != v2.Len() {
			return false
		}
		for _, k := range v1.MapKeys() {
			val1 := v1.MapIndex(k)
			val2 := v2.MapIndex(k)
			if !recur(val1, val2) {
				return false
			}
		}
		return true
	case goreflect.Func:
		if v1.IsNil() && v2.IsNil() {
			return true
		}
		return false
	default:
		return v1.Interface() == v2.Interface()
	}
}
