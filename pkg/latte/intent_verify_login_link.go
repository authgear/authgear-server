package latte

import (
	"context"
	"errors"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/otp"
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

		err := i.setSubmittedCode(deps, code)
		if errors.Is(err, otp.ErrCodeNotFound) || errors.Is(err, otp.ErrInvalidCode) {
			return nil, otp.ErrInvalidLoginLink
		} else if err != nil {
			return nil, err
		}

		return workflow.NewNodeSimple(
			&NodeVerifiedLoginLink{},
		), nil
	default:
		return nil, workflow.ErrIncompatibleInput
	}
}

func (i *IntentVerifyLoginLink) setSubmittedCode(deps *workflow.Dependencies, code string) error {
	kind := otp.KindOOBOTP(deps.Config, model.AuthenticatorOOBChannelEmail)

	target, err := deps.OTPCodes.LookupCode(kind, code)
	if err != nil {
		return err
	}

	err = deps.OTPCodes.VerifyOTP(kind, target, code, &otp.VerifyOptions{
		// No need pass user ID (for rate limit checking),
		// since able to lookup by code strongly implies valid request.
		SkipConsume: true,
	})
	if err != nil {
		return err
	}

	state, err := deps.OTPCodes.SetSubmittedCode(kind, target, code)
	if err != nil {
		return err
	}

	if state.WorkflowID != "" {
		err = deps.WorkflowEvents.Publish(state.WorkflowID, workflow.NewEventRefresh())
		if err != nil {
			return err
		}
	}

	return nil
}

func (i *IntentVerifyLoginLink) GetEffects(ctx context.Context, deps *workflow.Dependencies, w *workflow.Workflow) (effs []workflow.Effect, err error) {
	return nil, nil
}

func (i *IntentVerifyLoginLink) OutputData(ctx context.Context, deps *workflow.Dependencies, w *workflow.Workflow) (interface{}, error) {
	return map[string]interface{}{}, nil
}
