package workflow

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/authgear/authgear-server/pkg/util/validation"
)

// Intent can optionally implement CookieGetter.
type Intent interface {
	InputReactor
	EffectGetter
	Kind() string
	JSONSchema() *validation.SimpleSchema
	OutputData(ctx context.Context, deps *Dependencies, workflow *Workflow) (interface{}, error)
}

type IntentOutput struct {
	Kind string      `json:"kind"`
	Data interface{} `json:"data,omitempty"`
}

type IntentJSON struct {
	Kind string          `json:"kind"`
	Data json.RawMessage `json:"data"`
}

type IntentFactory func() Intent

var intentRegistry = map[string]IntentFactory{}

func RegisterIntent(intent Intent) {
	intentType := reflect.TypeOf(intent).Elem()

	intentKind := intent.Kind()
	factory := IntentFactory(func() Intent {
		return reflect.New(intentType).Interface().(Intent)
	})

	if _, hasKind := intentRegistry[intentKind]; hasKind {
		panic(fmt.Errorf("workflow: duplicated intent kind: %v", intentKind))
	}
	intentRegistry[intentKind] = factory
}

func InstantiateIntent(j IntentJSON) (Intent, error) {
	factory, ok := intentRegistry[j.Kind]
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
