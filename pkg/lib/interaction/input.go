package interaction

import (
	"reflect"
)

func Input(i interface{}, input interface{}) bool {
	val := reflect.ValueOf(input)
	typ := val.Type()
	targetType := typ.Elem()
	for {
		if reflect.TypeOf(i).AssignableTo(targetType) {
			val.Elem().Set(reflect.ValueOf(i))
			return true
		}
		if x, ok := i.(interface{ Input() interface{} }); ok {
			i = x.Input()
		} else {
			break
		}
	}
	return false
}
