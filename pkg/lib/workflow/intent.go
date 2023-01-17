package workflow

import (
	"context"
	"encoding/json"
	"reflect"
)

type Intent interface {
	GetEffects(ctx context.Context, deps *Dependencies, workflow *Workflow) (effs []Effect, err error)
	DeriveEdges(ctx context.Context, deps *Dependencies, workflow *Workflow) ([]Edge, error)
	OutputData(ctx context.Context, deps *Dependencies, workflow *Workflow) (interface{}, error)
}

type IntentOutput struct {
	Kind string      `json:"kind"`
	Data interface{} `json:"data,omitempty"`
}

type intentJSON struct {
	Kind string          `json:"kind"`
	Data json.RawMessage `json:"data"`
}

type IntentFactory func() Intent

var intentRegistry = map[string]IntentFactory{}

func RegisterIntent(intent Intent) {
	intentType := reflect.TypeOf(intent).Elem()

	intentKind := intentType.Name()
	factory := IntentFactory(func() Intent {
		return reflect.New(intentType).Interface().(Intent)
	})

	if _, hasKind := intentRegistry[intentKind]; hasKind {
		panic("interaction: duplicated intent kind: " + intentKind)
	}
	intentRegistry[intentKind] = factory
}

func IntentKind(intent Intent) string {
	intentType := reflect.TypeOf(intent).Elem()
	return intentType.Name()
}

func InstantiateIntent(kind string) Intent {
	factory, ok := intentRegistry[kind]
	if !ok {
		panic("interaction: unknown intent kind: " + kind)
	}
	return factory()
}
