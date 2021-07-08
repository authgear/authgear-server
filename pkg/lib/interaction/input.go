package interaction

import (
	"reflect"
)

func AsInput(i interface{}, input interface{}) bool {
	if i == nil {
		return false
	}
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
