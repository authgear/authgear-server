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
	OutputData(ctx context.Context, deps *Dependencies, workflows Workflows) (interface{}, error)
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

var publicIntentRegistry = map[string]IntentFactory{}
var privateIntentRegistry = map[string]IntentFactory{}

func RegisterPublicIntent(intent Intent) {
	intentType := reflect.TypeOf(intent).Elem()

	intentKind := intent.Kind()
	factory := IntentFactory(func() Intent {
		return reflect.New(intentType).Interface().(Intent)
	})

	if _, hasKind := publicIntentRegistry[intentKind]; hasKind {
		panic(fmt.Errorf("workflow: duplicated intent kind: %v", intentKind))
	}
	if _, hasKind := privateIntentRegistry[intentKind]; hasKind {
		panic(fmt.Errorf("workflow: duplicated intent kind: %v", intentKind))
	}

	publicIntentRegistry[intentKind] = factory
	privateIntentRegistry[intentKind] = factory
}

func RegisterPrivateIntent(intent Intent) {
	intentType := reflect.TypeOf(intent).Elem()

	intentKind := intent.Kind()
	factory := IntentFactory(func() Intent {
		return reflect.New(intentType).Interface().(Intent)
	})

	if _, hasKind := privateIntentRegistry[intentKind]; hasKind {
		panic(fmt.Errorf("workflow: duplicated intent kind: %v", intentKind))
	}

	privateIntentRegistry[intentKind] = factory
}

func InstantiateIntentFromPublicRegistry(ctx context.Context, j IntentJSON) (Intent, error) {
	factory, ok := publicIntentRegistry[j.Kind]
	if !ok {
		return nil, ErrUnknownIntent
	}
	intent := factory()
	err := intent.JSONSchema().Validator().ParseJSONRawMessage(ctx, j.Data, intent)
	if err != nil {
		return nil, err
	}

	return intent, nil
}

func InstantiateIntentFromPrivateRegistry(j IntentJSON) (Intent, error) {
	factory, ok := privateIntentRegistry[j.Kind]
	if !ok {
		return nil, ErrUnknownIntent
	}
	intent := factory()

	err := json.Unmarshal(j.Data, intent)
	if err != nil {
		return nil, err
	}

	return intent, nil
}
