package latte

import (
	"context"
	"time"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/otp"
	"github.com/authgear/authgear-server/pkg/lib/feature/verification"
	"github.com/authgear/authgear-server/pkg/lib/infra/mail"
	"github.com/authgear/authgear-server/pkg/lib/translation"
	"github.com/authgear/authgear-server/pkg/lib/workflow"
)

func init() {
	workflow.RegisterNode(&NodeVerifyEmail{})
}

var nodeVerifyEmailOTPForm = otp.FormCode

type NodeVerifyEmail struct {
	UserID     string `json:"user_id"`
	IdentityID string `json:"identity_id"`
	Email      string `json:"email"`
}

func (n *NodeVerifyEmail) Kind() string {
	return "latte.NodeVerifyEmail"
}

func (n *NodeVerifyEmail) GetEffects(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) (effs []workflow.Effect, err error) {
	return nil, nil
}

func (*NodeVerifyEmail) CanReactTo(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) ([]workflow.Input, error) {
	return []workflow.Input{
		&InputTakeOOBOTPCode{},
		&InputResendOOBOTPCode{},
	}, nil
}

func (n *NodeVerifyEmail) ReactTo(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows, input workflow.Input) (*workflow.Node, error) {
	var inputTakeOOBOTPCode inputTakeOOBOTPCode
	var inputResendOOBOTPCode inputResendOOBOTPCode

	switch {
	case workflow.AsInput(input, &inputTakeOOBOTPCode):
		code := inputTakeOOBOTPCode.GetCode()

		err := deps.OTPCodes.VerifyOTP(
			otp.KindVerification(deps.Config, model.AuthenticatorOOBChannelEmail),
			n.Email,
			code,
			&otp.VerifyOptions{UserID: n.UserID},
		)
		if apierrors.IsKind(err, otp.InvalidOTPCode) {
			return nil, verification.ErrInvalidVerificationCode
		} else if err != nil {
			return nil, err
		}

		verifiedClaim := deps.Verification.NewVerifiedClaim(n.UserID, string(model.ClaimEmail), n.Email)
		return workflow.NewNodeSimple(&NodeVerifiedIdentity{
			IdentityID:       n.IdentityID,
			NewVerifiedClaim: verifiedClaim,
		}), nil

	case workflow.AsInput(input, &inputResendOOBOTPCode):
		err := n.sendCode(ctx, deps)
		if err != nil {
			return nil, err
		}
		return workflow.NewNodeSimple(n), workflow.ErrUpdateNode

	default:
		return nil, workflow.ErrIncompatibleInput
	}
}

func (n *NodeVerifyEmail) OutputData(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) (interface{}, error) {
	target := n.Email
	state, err := deps.OTPCodes.InspectState(otp.KindVerification(deps.Config, model.AuthenticatorOOBChannelEmail), target)
	if err != nil {
		return nil, err
	}

	type NodeVerifyEmailOutput struct {
		MaskedEmail                    string    `json:"masked_email"`
		CodeLength                     int       `json:"code_length"`
		CanResendAt                    time.Time `json:"can_resend_at"`
		FailedAttemptRateLimitExceeded bool      `json:"failed_attempt_rate_limit_exceeded"`
	}

	return NodeVerifyEmailOutput{
		MaskedEmail:                    mail.MaskAddress(target),
		CodeLength:                     nodeVerifyEmailOTPForm.CodeLength(),
		CanResendAt:                    state.CanResendAt,
		FailedAttemptRateLimitExceeded: state.TooManyAttempts,
	}, nil
}

func (n *NodeVerifyEmail) otpKind(deps *workflow.Dependencies) otp.Kind {
	return otp.KindVerification(deps.Config, model.AuthenticatorOOBChannelEmail)
}

func (n *NodeVerifyEmail) otpTarget() string {
	return n.Email
}

func (n *NodeVerifyEmail) sendCode(ctx context.Context, deps *workflow.Dependencies) error {
	msg, err := deps.OTPSender.Prepare(model.AuthenticatorOOBChannelEmail, n.Email, nodeVerifyEmailOTPForm, translation.MessageTypeVerification)
	if err != nil {
		return err
	}
	defer msg.Close()

	code, err := deps.OTPCodes.GenerateOTP(
		n.otpKind(deps),
		n.Email,
		nodeVerifyEmailOTPForm,
		&otp.GenerateOptions{
			UserID:     n.UserID,
			WorkflowID: workflow.GetWorkflowID(ctx),
		},
	)
	if err != nil {
		return err
	}

	err = deps.OTPSender.Send(msg, otp.SendOptions{OTP: code})
	if err != nil {
		return err
	}

	return nil

}
