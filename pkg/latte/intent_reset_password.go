package latte

import (
	"context"

	"github.com/authgear/authgear-server/pkg/lib/workflow"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

func init() {
	workflow.RegisterPublicIntent(&IntentResetPassword{})
}

var IntentResetPasswordSchema = validation.NewSimpleSchema(`
	{
		"type": "object",
		"additionalProperties": false
	}
`)

type IntentResetPassword struct {
}

func (*IntentResetPassword) Kind() string {
	return "latte.IntentResetPassword"
}

func (*IntentResetPassword) JSONSchema() *validation.SimpleSchema {
	return IntentResetPasswordSchema
}

func (*IntentResetPassword) CanReactTo(ctx context.Context, deps *workflow.Dependencies, w *workflow.Workflow) ([]workflow.Input, error) {
	if len(w.Nodes) == 0 {
		return []workflow.Input{
			&InputResetPasswordByCode{},
		}, nil
	}
	return nil, workflow.ErrEOF
}

func (i *IntentResetPassword) ReactTo(ctx context.Context, deps *workflow.Dependencies, w *workflow.Workflow, input workflow.Input) (*workflow.Node, error) {
	var inputResetPasswordByCode inputResetPasswordByCode

	switch {
	case workflow.AsInput(input, &inputResetPasswordByCode):
		node := NodeDoResetPasswordByCode{
			Code:        inputResetPasswordByCode.GetCode(),
			NewPassword: inputResetPasswordByCode.GetNewPassword(),
		}
		return workflow.NewNodeSimple(&node), nil
	default:
		return nil, workflow.ErrIncompatibleInput
	}

}

func (*IntentResetPassword) GetEffects(ctx context.Context, deps *workflow.Dependencies, w *workflow.Workflow) (effs []workflow.Effect, err error) {
	return nil, nil
}

func (*IntentResetPassword) OutputData(ctx context.Context, deps *workflow.Dependencies, w *workflow.Workflow) (interface{}, error) {
	return nil, nil
}
