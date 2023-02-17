package latte

import (
	"context"

	"github.com/authgear/authgear-server/pkg/lib/workflow"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

func init() {
	workflow.RegisterPublicIntent(&IntentForgotPassword{})
}

var IntentForgotPasswordSchema = validation.NewSimpleSchema(`{}`)

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
		node.sendCode(ctx, deps, w)
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
