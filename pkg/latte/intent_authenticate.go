package latte

import (
	"context"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/workflow"
	"github.com/authgear/authgear-server/pkg/util/stringutil"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

func init() {
	workflow.RegisterPublicIntent(&IntentAuthenticate{})
}

var IntentAuthenticateSchema = validation.NewSimpleSchema(`
	{
		"type": "object",
		"additionalProperties": false
	}
`)

type IntentAuthenticate struct {
	CaptchaProtectedIntent
}

func (*IntentAuthenticate) Kind() string {
	return "latte.IntentAuthenticate"
}

func (*IntentAuthenticate) JSONSchema() *validation.SimpleSchema {
	return IntentAuthenticateSchema
}

func (*IntentAuthenticate) CanReactTo(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) ([]workflow.Input, error) {
	if len(workflows.Nearest.Nodes) == 0 {
		return []workflow.Input{
			&InputTakeLoginID{},
			&InputTakeProofOfPhoneNumberVerification{},
		}, nil
	}
	if len(workflows.Nearest.Nodes) == 1 && len(workflow.FindSubWorkflows[*IntentVerifyProofOfPhoneNumberVerification](workflows.Nearest)) != 0 {
		return []workflow.Input{
			&InputTakeProofOfPhoneNumberVerification{},
		}, nil
	} else {
		return nil, workflow.ErrEOF
	}
}

func (i *IntentAuthenticate) ReactTo(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows, input workflow.Input) (*workflow.Node, error) {
	var inputTakeLoginID inputTakeLoginID
	var inputTakeProofOfPhoneNumberVerification inputTakeProofOfPhoneNumberVerification

	switch len(workflows.Nearest.Nodes) {
	case 0:
		switch {
		case workflow.AsInput(input, &inputTakeLoginID):
			loginID := inputTakeLoginID.GetLoginID()
			spec := &identity.Spec{
				Type: model.IdentityTypeLoginID,
				LoginID: &identity.LoginIDSpec{
					Value: stringutil.NewUserInputString(loginID),
				},
			}

			// TODO: account enumeration? although need OTP to proceed, login/signup is indicated in workflow data.

			exactMatch, _, err := deps.Identities.SearchBySpec(ctx, spec)
			if err != nil {
				return nil, err
			}
			if exactMatch == nil {
				intent := &IntentSignup{}
				intent.IsCaptchaProtected = i.IsCaptchaProtected
				return workflow.NewSubWorkflow(intent), nil
			}
			intent := &IntentLogin{
				Identity: exactMatch,
			}
			intent.IsCaptchaProtected = i.IsCaptchaProtected
			return workflow.NewSubWorkflow(intent), nil

		case workflow.AsInput(input, &inputTakeProofOfPhoneNumberVerification):
			proofOfPhoneNumberVerification := inputTakeProofOfPhoneNumberVerification.GetProofOfPhoneNumberVerification()
			intent := &IntentVerifyProofOfPhoneNumberVerification{
				ProofOfPhoneNumberVerification: proofOfPhoneNumberVerification,
			}
			return workflow.NewSubWorkflow(intent), nil

		default:
			return nil, workflow.ErrIncompatibleInput
		}
	case 1:
		if workflow.AsInput(input, &inputTakeProofOfPhoneNumberVerification) {
			if ws := workflow.FindSubWorkflows[*IntentVerifyProofOfPhoneNumberVerification](workflows.Nearest); len(ws) > 0 {
				if n, found := workflow.FindSingleNode[*NodeVerifiedProofOfPhoneNumberVerification](ws[0]); found {
					identityInfo, _, err := deps.Identities.SearchBySpec(ctx, n.IdentitySpec)
					if err != nil {
						return nil, err
					}
					if identityInfo == nil {
						intent := &IntentSignup{
							PhoneNumberHint: n.IdentitySpec.LoginID.Value.TrimSpace(),
						}
						intent.IsCaptchaProtected = i.IsCaptchaProtected
						return workflow.NewSubWorkflow(intent), nil
					}

					intent := &IntentLogin{
						Identity: identityInfo,
					}
					intent.IsCaptchaProtected = i.IsCaptchaProtected
					return workflow.NewSubWorkflow(intent), nil
				}
			}
		}
		return nil, workflow.ErrIncompatibleInput
	}

	return nil, workflow.ErrIncompatibleInput
}

func (*IntentAuthenticate) GetEffects(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) (effs []workflow.Effect, err error) {
	// IntentAuthenticate has no effects.
	// The effects would be in IntentSignup or IntentLogin.
	return nil, nil
}

func (*IntentAuthenticate) OutputData(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) (interface{}, error) {
	return nil, nil
}
