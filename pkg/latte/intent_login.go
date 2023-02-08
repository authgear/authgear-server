package latte

import (
	"context"

	"github.com/authgear/authgear-server/pkg/lib/workflow"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

func init() {
	workflow.RegisterIntent(&IntentLogin{})
}

var IntentLoginSchema = validation.NewSimpleSchema(`
	{
		"type": "object",
		"additionalProperties": false,
		"properties": {
			"some_string": { "type": "string" }
		},
		"required": ["some_string"]
	}
`)

type IntentLogin struct {
	SomeString string `json:"some_string"`
}

func (*IntentLogin) Kind() string {
	return "latte.IntentLogin"
}

func (*IntentLogin) JSONSchema() *validation.SimpleSchema {
	return IntentLoginSchema
}

func (*IntentLogin) CanReactTo(ctx context.Context, deps *workflow.Dependencies, w *workflow.Workflow) ([]workflow.Input, error) {
	if len(w.Nodes) == 0 {
		return []workflow.Input{
			&InputTakeLoginID{},
		}, nil
	}
	return nil, workflow.ErrEOF
}

func (*IntentLogin) ReactTo(ctx context.Context, deps *workflow.Dependencies, w *workflow.Workflow, input workflow.Input) (*workflow.Node, error) {
	var loginIDInput inputTakeLoginID

	switch {
	case workflow.AsInput(input, &loginIDInput):
		return workflow.NewNodeSimple(&NodeTakeLoginID{
			LoginID: loginIDInput.GetLoginID(),
		}), nil
	default:
		return nil, workflow.ErrIncompatibleInput
	}
}

func (*IntentLogin) GetEffects(ctx context.Context, deps *workflow.Dependencies, w *workflow.Workflow) (effs []workflow.Effect, err error) {
	return nil, nil
}

func (i *IntentLogin) OutputData(ctx context.Context, deps *workflow.Dependencies, w *workflow.Workflow) (interface{}, error) {
	return map[string]interface{}{
		"some_string": i.SomeString,
	}, nil
}
