package latte

import (
	"context"

	"github.com/authgear/authgear-server/pkg/lib/workflow"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

func init() {
	workflow.RegisterPublicIntent(&IntentApproveLoginLink{})
}

var IntentApproveLoginLinkSchema = validation.NewSimpleSchema(`{}`)

type IntentApproveLoginLink struct{}

func (i *IntentApproveLoginLink) Kind() string {
	return "latte.IntentApproveLoginLink"
}

func (i *IntentApproveLoginLink) JSONSchema() *validation.SimpleSchema {
	return IntentApproveLoginLinkSchema
}

func (i *IntentApproveLoginLink) CanReactTo(ctx context.Context, deps *workflow.Dependencies, w *workflow.Workflow) ([]workflow.Input, error) {
	if len(w.Nodes) == 0 {
		return []workflow.Input{
			&InputTakeLoginLinkCode{},
		}, nil
	}

	return nil, workflow.ErrEOF
}

func (i *IntentApproveLoginLink) ReactTo(ctx context.Context, deps *workflow.Dependencies, w *workflow.Workflow, input workflow.Input) (*workflow.Node, error) {
	var inputTakeLoginLinkCode inputTakeLoginLinkCode

	switch {
	case workflow.AsInput(input, &inputTakeLoginLinkCode):
		code := inputTakeLoginLinkCode.GetCode()
		_, err := deps.OTPCodes.VerifyMagicLinkCode(code, false)
		if err != nil {
			return nil, err
		}

		return workflow.NewNodeSimple(
			&NodeVerifiedLoginLink{Code: code},
		), nil
	default:
		return nil, workflow.ErrIncompatibleInput
	}
}

func (i *IntentApproveLoginLink) GetEffects(ctx context.Context, deps *workflow.Dependencies, w *workflow.Workflow) (effs []workflow.Effect, err error) {
	return nil, nil
}

func (i *IntentApproveLoginLink) OutputData(ctx context.Context, deps *workflow.Dependencies, w *workflow.Workflow) (interface{}, error) {
	return map[string]interface{}{}, nil
}
