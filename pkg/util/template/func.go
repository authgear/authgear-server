package template

import (
	"fmt"
	"reflect"
)

// MakeMap creates a map with the given key value pairs.
func MakeMap(pairs ...interface{}) map[string]interface{} {
	length := len(pairs)
	if length%2 == 1 {
		panic(fmt.Errorf("template: the length of the argument must be even: %v", length))
	}

	out := make(map[string]interface{})
	for i := 0; i < length; i += 2 {
		key, ok := pairs[i].(string)
		if !ok {
			panic(fmt.Errorf("template: unexpected key type: %T %v", pairs[i], pairs[i]))
		}
		out[key] = pairs[i+1]
	}

	return out
}

// Add returns the sum of a and b.
func Add(b, a interface{}) interface{} {
	av := reflect.ValueOf(a)
	bv := reflect.ValueOf(b)

	switch av.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		switch bv.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return av.Int() + bv.Int()
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			return av.Int() + int64(bv.Uint())
		case reflect.Float32, reflect.Float64:
			return float64(av.Int()) + bv.Float()
		default:
			panic(fmt.Errorf("template: unknown type for %q (%T)", bv, b))
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		switch bv.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return int64(av.Uint()) + bv.Int()
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			return av.Uint() + bv.Uint()
		case reflect.Float32, reflect.Float64:
			return float64(av.Uint()) + bv.Float()
		default:
			panic(fmt.Errorf("template: unknown type for %q (%T)", bv, b))
		}
	case reflect.Float32, reflect.Float64:
		switch bv.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return av.Float() + float64(bv.Int())
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			return av.Float() + float64(bv.Uint())
		case reflect.Float32, reflect.Float64:
			return av.Float() + bv.Float()
		default:
			panic(fmt.Errorf("template: unknown type for %q (%T)", bv, b))
		}
	default:
		panic(fmt.Errorf("template: unknown type for %q (%T)", av, a))
	}
}
