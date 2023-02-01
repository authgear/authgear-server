package workflow

import (
	"fmt"
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

func InstantiateInput(kind string) (Input, error) {
	factory, ok := inputRegistry[kind]
	if !ok {
		return nil, fmt.Errorf("workflow: unknown input kind: %v", kind)
	}
	return factory(), nil
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
