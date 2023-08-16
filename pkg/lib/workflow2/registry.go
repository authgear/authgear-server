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

type IntentJSON struct {
	Kind string          `json:"kind"`
	Data json.RawMessage `json:"data"`
}

type InputJSON struct {
	Kind string          `json:"kind"`
	Data json.RawMessage `json:"data"`
}

type publicIntentFactory func() PublicIntent

type inputFactory func() Input

var publicIntentRegistry = map[string]publicIntentFactory{}
var publicInputRegistry = map[string]inputFactory{}

func RegisterPublicIntent(intent PublicIntent) {
	intentType := reflect.TypeOf(intent).Elem()

	intentKind := intent.Kind()
	factory := publicIntentFactory(func() PublicIntent {
		return reflect.New(intentType).Interface().(PublicIntent)
	})

	if _, hasKind := publicIntentRegistry[intentKind]; hasKind {
		panic(fmt.Errorf("workflow: duplicated intent kind: %v", intentKind))
	}

	publicIntentRegistry[intentKind] = factory

	RegisterIntent(intent)
}

func InstantiateIntentFromPublicRegistry(j IntentJSON) (PublicIntent, error) {
	factory, ok := publicIntentRegistry[j.Kind]
	if !ok {
		return nil, ErrUnknownIntent
	}
	intent := factory()

	err := intent.JSONSchema().Validator().ParseJSONRawMessage(j.Data, intent)
	if err != nil {
		return nil, err
	}

	return intent, nil
}

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
