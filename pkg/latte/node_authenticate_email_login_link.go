package latte

import (
	"context"
	"time"

	"github.com/authgear/authgear-server/pkg/api"
	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/authn/otp"
	"github.com/authgear/authgear-server/pkg/lib/infra/mail"
	"github.com/authgear/authgear-server/pkg/lib/workflow"
)

func init() {
	workflow.RegisterNode(&NodeAuthenticateEmailLoginLink{})
}

type NodeAuthenticateEmailLoginLink struct {
	Authenticator *authenticator.Info `json:"authenticator,omitempty"`
}

func (n *NodeAuthenticateEmailLoginLink) Kind() string {
	return "latte.NodeAuthenticateEmailLoginLink"
}

func (n *NodeAuthenticateEmailLoginLink) GetEffects(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) (effs []workflow.Effect, err error) {
	return nil, nil
}

func (n *NodeAuthenticateEmailLoginLink) CanReactTo(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) ([]workflow.Input, error) {
	return []workflow.Input{
		&InputCheckLoginLinkVerified{},
		&InputResendOOBOTPCode{},
	}, nil
}

func (n *NodeAuthenticateEmailLoginLink) ReactTo(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows, input workflow.Input) (*workflow.Node, error) {
	var inputCheckLoginLinkVerified inputCheckLoginLinkVerified
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
			OTPForm:           otp.FormLink,
			IsResend:          true,
		}).Do()
		if err != nil {
			return nil, err
		}

		return nil, workflow.ErrSameNode
	case workflow.AsInput(input, &inputCheckLoginLinkVerified):
		info := n.Authenticator
		err := deps.OTPCodes.VerifyOTP(
			otp.KindOOBOTPLink(deps.Config, model.AuthenticatorOOBChannelEmail),
			info.OOBOTP.Email,
			"",
			&otp.VerifyOptions{
				UseSubmittedCode: true,
				UserID:           info.UserID,
			},
		)
		if apierrors.IsKind(err, otp.InvalidOTPCode) {
			// Don't fire the AuthenticationFailedEvent
			// The event should be fired only when the user submits code through the login link
			return nil, api.ErrInvalidCredentials
		} else if err != nil {
			return nil, err
		}
		return workflow.NewNodeSimple(&NodeVerifiedAuthenticator{
			Authenticator: info,
		}), nil
	}
	return nil, workflow.ErrIncompatibleInput
}

func (n *NodeAuthenticateEmailLoginLink) OutputData(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) (interface{}, error) {
	target := n.Authenticator.OOBOTP.Email
	state, err := deps.OTPCodes.InspectState(
		otp.KindOOBOTPLink(deps.Config, model.AuthenticatorOOBChannelEmail),
		target,
	)
	if err != nil {
		return nil, err
	}

	type NodeAuthenticateEmailLoginLinkOutput struct {
		LoginLinkSubmitted             bool      `json:"login_link_submitted"`
		MaskedEmail                    string    `json:"masked_email"`
		CanResendAt                    time.Time `json:"can_resend_at"`
		FailedAttemptRateLimitExceeded bool      `json:"failed_attempt_rate_limit_exceeded"`
	}

	return NodeAuthenticateEmailLoginLinkOutput{
		LoginLinkSubmitted:             state.SubmittedCode != "",
		MaskedEmail:                    mail.MaskAddress(target),
		CanResendAt:                    state.CanResendAt,
		FailedAttemptRateLimitExceeded: state.TooManyAttempts,
	}, nil
}
