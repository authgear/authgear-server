package latte

import (
	"context"

	"github.com/authgear/authgear-server/pkg/lib/workflow"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

func init() {
	workflow.RegisterPublicIntent(&IntentVerifyLoginLink{})
}

var IntentVerifyLoginLinkSchema = validation.NewSimpleSchema(`{}`)

type IntentVerifyLoginLink struct{}

func (i *IntentVerifyLoginLink) Kind() string {
	return "latte.IntentVerifyLoginLink"
}

func (i *IntentVerifyLoginLink) JSONSchema() *validation.SimpleSchema {
	return IntentVerifyLoginLinkSchema
}

func (i *IntentVerifyLoginLink) CanReactTo(ctx context.Context, deps *workflow.Dependencies, w *workflow.Workflow) ([]workflow.Input, error) {
	if len(w.Nodes) == 0 {
		return []workflow.Input{
			&InputTakeLoginLinkCode{},
		}, nil
	}

	return nil, workflow.ErrEOF
}

func (i *IntentVerifyLoginLink) ReactTo(ctx context.Context, deps *workflow.Dependencies, w *workflow.Workflow, input workflow.Input) (*workflow.Node, error) {
	var inputTakeLoginLinkCode inputTakeLoginLinkCode

	switch {
	case workflow.AsInput(input, &inputTakeLoginLinkCode):
		code := inputTakeLoginLinkCode.GetCode()
		codeModal, err := deps.OTPCodes.SetUserInputtedLoginLinkCode(code)
		if err != nil {
			return nil, err
		}

		if codeModal.WorkflowID != "" {
			err = deps.WorkflowEvents.Publish(codeModal.WorkflowID, workflow.NewEventRefresh())
			if err != nil {
				return nil, err
			}
		}

		return workflow.NewNodeSimple(
			&NodeVerifiedLoginLink{},
		), nil
	default:
		return nil, workflow.ErrIncompatibleInput
	}
}

func (i *IntentVerifyLoginLink) GetEffects(ctx context.Context, deps *workflow.Dependencies, w *workflow.Workflow) (effs []workflow.Effect, err error) {
	return nil, nil
}

func (i *IntentVerifyLoginLink) OutputData(ctx context.Context, deps *workflow.Dependencies, w *workflow.Workflow) (interface{}, error) {
	return map[string]interface{}{}, nil
}
