package latte

import (
	"context"
	"time"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/authn/otp"
	"github.com/authgear/authgear-server/pkg/lib/facade"
	"github.com/authgear/authgear-server/pkg/lib/workflow"
	"github.com/authgear/authgear-server/pkg/util/phone"
)

func init() {
	workflow.RegisterNode(&NodeAuthenticateOOBOTPPhone{})
}

type NodeAuthenticateOOBOTPPhone struct {
	Authenticator *authenticator.Info `json:"authenticator,omitempty"`
}

func (n *NodeAuthenticateOOBOTPPhone) Kind() string {
	return "latte.NodeAuthenticateOOBOTPPhone"
}

func (n *NodeAuthenticateOOBOTPPhone) GetEffects(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) (effs []workflow.Effect, err error) {
	return nil, nil
}

func (n *NodeAuthenticateOOBOTPPhone) CanReactTo(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) ([]workflow.Input, error) {
	return []workflow.Input{
		&InputTakeOOBOTPCode{},
		&InputResendOOBOTPCode{},
	}, nil
}

func (n *NodeAuthenticateOOBOTPPhone) ReactTo(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows, input workflow.Input) (*workflow.Node, error) {
	var inputTakeOOBOTPCode inputTakeOOBOTPCode
	var inputResendOOBOTPCode inputResendOOBOTPCode
	switch {
	case workflow.AsInput(input, &inputResendOOBOTPCode):
		info := n.Authenticator
		err := (&SendOOBCode{
			WorkflowID:        workflow.GetWorkflowID(ctx),
			Deps:              deps,
			Stage:             authenticatorKindToStage(info.Kind),
			IsAuthenticating:  true,
			AuthenticatorInfo: info,
			OTPForm:           otp.FormCode,
			IsResend:          true,
		}).Do()
		if err != nil {
			return nil, err
		}

		return nil, workflow.ErrSameNode
	case workflow.AsInput(input, &inputTakeOOBOTPCode):
		info := n.Authenticator
		_, err := deps.Authenticators.VerifyWithSpec(info, &authenticator.Spec{
			OOBOTP: &authenticator.OOBOTPSpec{
				Code: inputTakeOOBOTPCode.GetCode(),
			},
		}, &facade.VerifyOptions{
			Form: otp.FormCode,
			AuthenticationDetails: facade.NewAuthenticationDetails(
				info.UserID,
				authn.AuthenticationStagePrimary,
				authn.AuthenticationTypeOOBOTPSMS,
			),
		})
		if err != nil {
			return nil, err
		}
		return workflow.NewNodeSimple(&NodeVerifiedAuthenticator{
			Authenticator: info,
		}), nil
	}
	return nil, workflow.ErrIncompatibleInput
}

func (n *NodeAuthenticateOOBOTPPhone) OutputData(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) (interface{}, error) {
	target := n.Authenticator.OOBOTP.Phone
	state, err := deps.OTPCodes.InspectState(
		otp.KindOOBOTPCode(deps.Config, model.AuthenticatorOOBChannelSMS),
		target,
	)
	if err != nil {
		return nil, err
	}

	type NodeAuthenticateOOBOTPPhoneOutput struct {
		MaskedPhoneNumber              string    `json:"masked_phone_number"`
		CanResendAt                    time.Time `json:"can_resend_at"`
		FailedAttemptRateLimitExceeded bool      `json:"failed_attempt_rate_limit_exceeded"`
	}

	return NodeAuthenticateOOBOTPPhoneOutput{
		MaskedPhoneNumber:              phone.Mask(target),
		CanResendAt:                    state.CanResendAt,
		FailedAttemptRateLimitExceeded: state.TooManyAttempts,
	}, nil
}
