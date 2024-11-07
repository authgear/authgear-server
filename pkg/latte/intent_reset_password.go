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

func (*IntentResetPassword) CanReactTo(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) ([]workflow.Input, error) {
	switch len(workflows.Nearest.Nodes) {
	case 0:
		return []workflow.Input{
			&InputTakeCode{},
		}, nil
	case 1:
		return []workflow.Input{
			&InputTakePassword{},
		}, nil
	}
	return nil, workflow.ErrEOF
}

func (i *IntentResetPassword) ReactTo(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows, input workflow.Input) (*workflow.Node, error) {
	switch len(workflows.Nearest.Nodes) {
	case 0:
		var inputTakeCode inputTakeCode
		if workflow.AsInput(input, &inputTakeCode) {
			state, err := deps.ResetPassword.VerifyCode(ctx, inputTakeCode.GetCode())
			if err != nil {
				return nil, err
			}
			node := NodeValidatedResetPasswordCode{
				Code:   inputTakeCode.GetCode(),
				UserID: state.UserID,
			}
			return workflow.NewNodeSimple(&node), nil
		}
	case 1:
		var inputTakeNewPassword inputTakeNewPassword
		var code = i.getValidatedCode(workflows.Nearest)
		if workflow.AsInput(input, &inputTakeNewPassword) {
			node := NodeDoResetPasswordByCode{
				Code:        code,
				NewPassword: inputTakeNewPassword.GetNewPassword(),
			}
			return workflow.NewNodeSimple(&node), nil
		}

	}
	return nil, workflow.ErrIncompatibleInput
}

func (*IntentResetPassword) GetEffects(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) (effs []workflow.Effect, err error) {
	return nil, nil
}

func (*IntentResetPassword) OutputData(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) (interface{}, error) {
	return nil, nil
}

func (*IntentResetPassword) getValidatedCode(w *workflow.Workflow) string {
	node, ok := workflow.FindSingleNode[*NodeValidatedResetPasswordCode](w)
	if !ok {
		return ""
	}
	return node.Code
}
