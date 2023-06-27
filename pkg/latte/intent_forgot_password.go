package latte

import (
	"context"
	"errors"

	"github.com/authgear/authgear-server/pkg/lib/feature/forgotpassword"
	"github.com/authgear/authgear-server/pkg/lib/workflow"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

func init() {
	workflow.RegisterPublicIntent(&IntentForgotPassword{})
}

var IntentForgotPasswordSchema = validation.NewSimpleSchema(`
	{
		"type": "object",
		"additionalProperties": false
	}
`)

type IntentForgotPassword struct {
}

func (*IntentForgotPassword) Kind() string {
	return "latte.IntentForgotPassword"
}

func (*IntentForgotPassword) JSONSchema() *validation.SimpleSchema {
	return IntentForgotPasswordSchema
}

func (*IntentForgotPassword) CanReactTo(ctx context.Context, deps *workflow.Dependencies, w *workflow.Workflow) ([]workflow.Input, error) {
	if len(w.Nodes) == 0 {
		return []workflow.Input{
			&InputTakeLoginID{},
		}, nil
	}
	return nil, workflow.ErrEOF
}

func (i *IntentForgotPassword) ReactTo(ctx context.Context, deps *workflow.Dependencies, w *workflow.Workflow, input workflow.Input) (*workflow.Node, error) {
	var inputTakeLoginID inputTakeLoginID

	switch {
	case workflow.AsInput(input, &inputTakeLoginID):
		loginID := inputTakeLoginID.GetLoginID()
		node := NodeSendForgotPasswordCode{LoginID: loginID}
		err := node.sendCode(ctx, deps, w)
		// We do not tell the user if the login ID was found
		if err != nil && !errors.Is(err, forgotpassword.ErrUserNotFound) {
			return nil, err
		}
		// From here, err == nil or errors.Is(err, forgotpassword.ErrUserNotFound)
		return workflow.NewNodeSimple(&node), nil
	default:
		return nil, workflow.ErrIncompatibleInput
	}
}

func (*IntentForgotPassword) GetEffects(ctx context.Context, deps *workflow.Dependencies, w *workflow.Workflow) (effs []workflow.Effect, err error) {
	return nil, nil
}

func (*IntentForgotPassword) OutputData(ctx context.Context, deps *workflow.Dependencies, w *workflow.Workflow) (interface{}, error) {
	return nil, nil
}
