package workflow

import (
	"reflect"
)

type Input interface {
	Kind() string
}

type InputFactory func() Input

var inputRegistry = map[string]InputFactory{}

func RegisterInput(input Input) {
	inputType := reflect.TypeOf(input).Elem()

	inputKind := input.Kind()
	factory := InputFactory(func() Input {
		return reflect.New(inputType).Interface().(Input)
	})

	if _, hasKind := inputRegistry[inputKind]; hasKind {
		panic("interaction: duplicated input kind: " + inputKind)
	}
	inputRegistry[inputKind] = factory
}

func InstantiateInput(kind string) Input {
	factory, ok := inputRegistry[kind]
	if !ok {
		panic("interaction: unknown input kind: " + kind)
	}
	return factory()
}

func AsInput(i Input, iface interface{}) bool {
	if i == nil {
		return false
	}
	val := reflect.ValueOf(iface)
	typ := val.Type()
	targetType := typ.Elem()
	for {
		if reflect.TypeOf(i).AssignableTo(targetType) {
			val.Elem().Set(reflect.ValueOf(i))
			return true
		}
		if x, ok := i.(interface{ Input() Input }); ok {
			i = x.Input()
		} else {
			break
		}
	}
	return false
}
