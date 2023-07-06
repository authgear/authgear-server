package latte

import (
	"context"

	"github.com/authgear/authgear-server/pkg/lib/workflow"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

func init() {
	workflow.RegisterPublicIntent(&IntentProtectedAuthenticate{})
}

var IntentProtectedAuthenticateSchema = validation.NewSimpleSchema(`
	{
		"type": "object",
		"additionalProperties": false
	}
`)

type IntentProtectedAuthenticate struct{}

func (*IntentProtectedAuthenticate) Kind() string {
	return "latte.IntentProtectedAuthenticate"
}

func (*IntentProtectedAuthenticate) JSONSchema() *validation.SimpleSchema {
	return IntentProtectedAuthenticateSchema
}

func (*IntentProtectedAuthenticate) CanReactTo(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) ([]workflow.Input, error) {
	if len(workflows.Nearest.Nodes) == 0 {
		return nil, nil
	}
	return nil, workflow.ErrEOF
}

func (*IntentProtectedAuthenticate) ReactTo(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows, input workflow.Input) (*workflow.Node, error) {
	if len(workflows.Nearest.Nodes) == 0 {
		intent := &IntentAuthenticate{}
		intent.IsCaptchaProtected = true
		return workflow.NewSubWorkflow(intent), nil
	}

	return nil, workflow.ErrIncompatibleInput
}

func (*IntentProtectedAuthenticate) GetEffects(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) (effs []workflow.Effect, err error) {
	return nil, nil
}

func (*IntentProtectedAuthenticate) OutputData(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) (interface{}, error) {
	return nil, nil
}
