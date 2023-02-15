package latte

import (
	"context"
	"errors"
	"time"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/otp"
	"github.com/authgear/authgear-server/pkg/lib/feature"
	"github.com/authgear/authgear-server/pkg/lib/feature/verification"
	"github.com/authgear/authgear-server/pkg/lib/ratelimit"
	"github.com/authgear/authgear-server/pkg/lib/workflow"
)

func init() {
	workflow.RegisterNode(&NodeVerifyPhoneSMS{})
}

type NodeVerifyPhoneSMS struct {
	UserID      string `json:"user_id"`
	IdentityID  string `json:"identity_id"`
	PhoneNumber string `json:"phone_number"`

	CodeLength int `json:"code_length"`
}

func (n *NodeVerifyPhoneSMS) Kind() string {
	return "latte.NodeVerifyPhoneSMS"
}

func (n *NodeVerifyPhoneSMS) GetEffects(ctx context.Context, deps *workflow.Dependencies, w *workflow.Workflow) (effs []workflow.Effect, err error) {
	return nil, nil
}

func (*NodeVerifyPhoneSMS) CanReactTo(ctx context.Context, deps *workflow.Dependencies, w *workflow.Workflow) ([]workflow.Input, error) {
	return []workflow.Input{
		&InputTakeOOBOTPCode{},
		&InputResendOOBOTPCode{},
	}, nil
}

func (n *NodeVerifyPhoneSMS) ReactTo(ctx context.Context, deps *workflow.Dependencies, w *workflow.Workflow, input workflow.Input) (*workflow.Node, error) {
	var inputTakeOOBOTPCode inputTakeOOBOTPCode
	var inputResendOOBOTPCode inputResendOOBOTPCode

	switch {
	case workflow.AsInput(input, &inputTakeOOBOTPCode):
		code := inputTakeOOBOTPCode.GetCode()

		err := deps.RateLimiter.TakeToken(verification.AutiBruteForceVerifyBucket(string(deps.RemoteIP)))
		if err != nil {
			return nil, err
		}

		err = deps.OTPCodes.VerifyCode(n.PhoneNumber, code)
		if errors.Is(err, otp.ErrInvalidCode) {
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
		err := n.sendCode(deps, w)
		if err != nil {
			return nil, err
		}
		return workflow.NewNodeSimple(n), workflow.ErrUpdateNode

	default:
		return nil, workflow.ErrIncompatibleInput
	}
}

func (n *NodeVerifyPhoneSMS) OutputData(ctx context.Context, deps *workflow.Dependencies, w *workflow.Workflow) (interface{}, error) {
	_, resetDuration, err := deps.RateLimiter.CheckToken(n.bucket(deps))
	if err != nil {
		return nil, err
	}

	now := deps.Clock.NowUTC()
	canResendAt := now.Add(resetDuration)

	type NodeVerifyPhoneNumberOutput struct {
		PhoneNumber string    `json:"phone_number"`
		CodeLength  int       `json:"code_length"`
		CanResendAt time.Time `json:"can_resend_at"`
	}

	return NodeVerifyPhoneNumberOutput{
		PhoneNumber: n.PhoneNumber,
		CodeLength:  n.CodeLength,
		CanResendAt: canResendAt,
	}, nil
}

func (n *NodeVerifyPhoneSMS) bucket(deps *workflow.Dependencies) ratelimit.Bucket {
	return AntiSpamSMSOTPCodeBucket(deps.Config.Messaging.SMS, n.PhoneNumber)
}

func (n *NodeVerifyPhoneSMS) sendCode(deps *workflow.Dependencies, w *workflow.Workflow) error {
	// disallow sending sms verification code if phone identity is disabled
	fc := deps.FeatureConfig
	if fc.Identity.LoginID.Types.Phone.Disabled {
		return feature.ErrFeatureDisabledSendingSMS
	}

	err := deps.RateLimiter.TakeToken(n.bucket(deps))
	if err != nil {
		return err
	}

	// FIXME: web session ID?
	code, err := deps.OTPCodes.GenerateCode(n.PhoneNumber, otp.OTPModeCode, "")
	if err != nil {
		return err
	}
	n.CodeLength = len(code.Code)

	err = deps.OOBCodeSender.SendCode(model.AuthenticatorOOBChannelSMS, n.PhoneNumber, code.Code, otp.MessageTypeVerification, otp.OTPModeCode)
	if err != nil {
		return err
	}

	return nil

}
