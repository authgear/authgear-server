package workflow

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/authgear/authgear-server/pkg/util/validation"
)

type InputReactor interface {
	CanReactTo(ctx context.Context, deps *Dependencies, workflows Workflows) ([]Input, error)
	ReactTo(ctx context.Context, deps *Dependencies, workflows Workflows, input Input) (*Node, error)
}

type Input interface {
	Kind() string
	JSONSchema() *validation.SimpleSchema
}

type InputJSON struct {
	Kind string          `json:"kind"`
	Data json.RawMessage `json:"data"`
}

// There is no private input factory because
// input cannot be deserialized from the database (at the moment).

type InputFactory func() Input

var publicInputRegistry = map[string]InputFactory{}

func RegisterPublicInput(input Input) {
	inputType := reflect.TypeOf(input).Elem()

	inputKind := input.Kind()
	factory := InputFactory(func() Input {
		return reflect.New(inputType).Interface().(Input)
	})

	if _, hasKind := publicInputRegistry[inputKind]; hasKind {
		panic(fmt.Errorf("workflow: duplicated input kind: %v", inputKind))
	}
	publicInputRegistry[inputKind] = factory
}

func InstantiateInputFromPublicRegistry(ctx context.Context, j InputJSON) (Input, error) {
	factory, ok := publicInputRegistry[j.Kind]
	if !ok {
		return nil, ErrUnknownInput
	}
	input := factory()
	err := input.JSONSchema().Validator().ParseJSONRawMessage(ctx, j.Data, input)
	if err != nil {
		return nil, err
	}

	return input, nil
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
