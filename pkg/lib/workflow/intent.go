package workflow

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"

	"github.com/authgear/authgear-server/pkg/util/validation"
)

type Intent interface {
	InputReactor
	Kind() string
	JSONSchema() *validation.SimpleSchema
	GetEffects(ctx context.Context, deps *Dependencies, workflow *Workflow) (effs []Effect, err error)
	OutputData(ctx context.Context, deps *Dependencies, workflow *Workflow) (interface{}, error)
}

type IntentCookieGetter interface {
	GetCookies(ctx context.Context, deps *Dependencies, workflow *Workflow) ([]*http.Cookie, error)
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
