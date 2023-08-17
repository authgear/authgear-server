package workflow2

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/authgear/authgear-server/pkg/util/validation"
)

type JSONSchemaGetter interface {
	JSONSchema() *validation.SimpleSchema
}

type InputJSON struct {
	Kind string          `json:"kind"`
	Data json.RawMessage `json:"data"`
}

type inputFactory func() Input

var publicInputRegistry = map[string]inputFactory{}

func RegisterPublicInput(input Input) {
	inputType := reflect.TypeOf(input).Elem()

	inputKind := input.Kind()
	factory := inputFactory(func() Input {
		return reflect.New(inputType).Interface().(Input)
	})

	if _, hasKind := publicInputRegistry[inputKind]; hasKind {
		panic(fmt.Errorf("workflow: duplicated input kind: %v", inputKind))
	}
	publicInputRegistry[inputKind] = factory
}

func InstantiateInputFromPublicRegistry(j InputJSON) (Input, error) {
	factory, ok := publicInputRegistry[j.Kind]
	if !ok {
		return nil, ErrUnknownInput
	}
	input := factory()

	err := input.JSONSchema().Validator().ParseJSONRawMessage(j.Data, input)
	if err != nil {
		return nil, err
	}

	return input, nil
}
