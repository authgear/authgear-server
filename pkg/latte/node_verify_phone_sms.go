package latte

import (
	"context"
	"time"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/otp"
	"github.com/authgear/authgear-server/pkg/lib/feature"
	"github.com/authgear/authgear-server/pkg/lib/feature/verification"
	"github.com/authgear/authgear-server/pkg/lib/translation"
	"github.com/authgear/authgear-server/pkg/lib/workflow"
	"github.com/authgear/authgear-server/pkg/util/phone"
)

func init() {
	workflow.RegisterNode(&NodeVerifyPhoneSMS{})
}

var nodeVerifyPhoneSMSForm = otp.FormCode

type NodeVerifyPhoneSMS struct {
	UserID      string `json:"user_id"`
	IdentityID  string `json:"identity_id"`
	PhoneNumber string `json:"phone_number"`
}

func (n *NodeVerifyPhoneSMS) Kind() string {
	return "latte.NodeVerifyPhoneSMS"
}

func (n *NodeVerifyPhoneSMS) GetEffects(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) (effs []workflow.Effect, err error) {
	return nil, nil
}

func (*NodeVerifyPhoneSMS) CanReactTo(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) ([]workflow.Input, error) {
	return []workflow.Input{
		&InputTakeOOBOTPCode{},
		&InputResendOOBOTPCode{},
	}, nil
}

func (n *NodeVerifyPhoneSMS) ReactTo(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows, input workflow.Input) (*workflow.Node, error) {
	var inputTakeOOBOTPCode inputTakeOOBOTPCode
	var inputResendOOBOTPCode inputResendOOBOTPCode

	switch {
	case workflow.AsInput(input, &inputTakeOOBOTPCode):
		code := inputTakeOOBOTPCode.GetCode()

		err := deps.OTPCodes.VerifyOTP(
			otp.KindVerification(deps.Config, model.AuthenticatorOOBChannelSMS),
			n.PhoneNumber,
			code,
			&otp.VerifyOptions{UserID: n.UserID},
		)
		if apierrors.IsKind(err, otp.InvalidOTPCode) {
			return nil, verification.ErrInvalidVerificationCode
		} else if err != nil {
			return nil, err
		}

		verifiedClaim := deps.Verification.NewVerifiedClaim(n.UserID, string(model.ClaimPhoneNumber), n.PhoneNumber)
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

func (n *NodeVerifyPhoneSMS) OutputData(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) (interface{}, error) {
	target := n.PhoneNumber
	state, err := deps.OTPCodes.InspectState(otp.KindVerification(deps.Config, model.AuthenticatorOOBChannelSMS), target)
	if err != nil {
		return nil, err
	}

	type NodeVerifyPhoneNumberOutput struct {
		MaskedPhoneNumber              string    `json:"masked_phone_number"`
		CodeLength                     int       `json:"code_length"`
		CanResendAt                    time.Time `json:"can_resend_at"`
		FailedAttemptRateLimitExceeded bool      `json:"failed_attempt_rate_limit_exceeded"`
	}

	return NodeVerifyPhoneNumberOutput{
		MaskedPhoneNumber:              phone.Mask(target),
		CodeLength:                     nodeVerifyPhoneSMSForm.CodeLength(),
		CanResendAt:                    state.CanResendAt,
		FailedAttemptRateLimitExceeded: state.TooManyAttempts,
	}, nil
}

func (n *NodeVerifyPhoneSMS) otpKind(deps *workflow.Dependencies) otp.Kind {
	return otp.KindVerification(deps.Config, model.AuthenticatorOOBChannelSMS)
}

func (n *NodeVerifyPhoneSMS) otpTarget() string {
	return n.PhoneNumber
}

func (n *NodeVerifyPhoneSMS) sendCode(ctx context.Context, deps *workflow.Dependencies) error {
	// disallow sending sms verification code if phone identity is disabled
	fc := deps.FeatureConfig
	if fc.Identity.LoginID.Types.Phone.Disabled {
		return feature.ErrFeatureDisabledSendingSMS
	}

	msg, err := deps.OTPSender.Prepare(model.AuthenticatorOOBChannelSMS, n.PhoneNumber, nodeVerifyPhoneSMSForm, translation.MessageTypeVerification)
	if err != nil {
		return err
	}
	defer msg.Close()

	code, err := deps.OTPCodes.GenerateOTP(
		n.otpKind(deps),
		n.PhoneNumber,
		nodeVerifyPhoneSMSForm,
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
