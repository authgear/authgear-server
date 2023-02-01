package latte

import (
	"context"

	"github.com/authgear/authgear-server/pkg/lib/workflow"
)

func init() {
	workflow.RegisterIntent(&IntentLogin{})
}

type IntentLogin struct {
	SomeString string `json:"some_string"`
}

func (*IntentLogin) Kind() string {
	return "latte.IntentLogin"
}

func (*IntentLogin) DeriveEdges(ctx context.Context, deps *workflow.Dependencies, w *workflow.Workflow) ([]workflow.Edge, error) {
	return nil, workflow.ErrEOF
}

func (*IntentLogin) GetEffects(ctx context.Context, deps *workflow.Dependencies, w *workflow.Workflow) (effs []workflow.Effect, err error) {
	return nil, nil
}

func (i *IntentLogin) OutputData(ctx context.Context, deps *workflow.Dependencies, w *workflow.Workflow) (interface{}, error) {
	return map[string]interface{}{
		"some_string": i.SomeString,
	}, nil
}
