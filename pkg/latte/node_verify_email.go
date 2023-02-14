package latte

import (
	"context"
	"errors"
	"time"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/otp"
	"github.com/authgear/authgear-server/pkg/lib/feature/verification"
	"github.com/authgear/authgear-server/pkg/lib/ratelimit"
	"github.com/authgear/authgear-server/pkg/lib/workflow"
)

func init() {
	workflow.RegisterNode(&NodeVerifyEmail{})
}

type NodeVerifyEmail struct {
	UserID     string `json:"user_id"`
	IdentityID string `json:"identity_id"`
	Email      string `json:"email"`

	CodeLength int `json:"code_length"`
}

func (n *NodeVerifyEmail) Kind() string {
	return "latte.NodeVerifyEmail"
}

func (n *NodeVerifyEmail) GetEffects(ctx context.Context, deps *workflow.Dependencies, w *workflow.Workflow) (effs []workflow.Effect, err error) {
	return nil, nil
}

func (*NodeVerifyEmail) CanReactTo(ctx context.Context, deps *workflow.Dependencies, w *workflow.Workflow) ([]workflow.Input, error) {
	return []workflow.Input{
		&InputTakeOOBOTPCode{},
		&InputResendCode{},
	}, nil
}

func (n *NodeVerifyEmail) ReactTo(ctx context.Context, deps *workflow.Dependencies, w *workflow.Workflow, input workflow.Input) (*workflow.Node, error) {
	var inputTakeOOBOTPCode inputTakeOOBOTPCode
	var inputResendCode inputResendCode

	switch {
	case workflow.AsInput(input, &inputTakeOOBOTPCode):
		code := inputTakeOOBOTPCode.GetCode()

		err := deps.RateLimiter.TakeToken(verification.AutiBruteForceVerifyBucket(string(deps.RemoteIP)))
		if err != nil {
			return nil, err
		}

		err = deps.OTPCodes.VerifyCode(n.Email, code)
		if errors.Is(err, otp.ErrInvalidCode) {
			return nil, verification.ErrInvalidVerificationCode
		} else if err != nil {
			return nil, err
		}

		verifiedClaim := deps.Verification.NewVerifiedClaim(n.UserID, string(model.ClaimEmail), n.Email)
		return workflow.NewNodeSimple(&NodeVerifiedIdentity{
			IdentityID:       n.IdentityID,
			NewVerifiedClaim: verifiedClaim,
		}), nil

	case workflow.AsInput(input, &inputResendCode):
		err := n.sendCode(deps, w)
		if err != nil {
			return nil, err
		}
		return workflow.NewNodeSimple(n), workflow.ErrUpdateNode

	default:
		return nil, workflow.ErrIncompatibleInput
	}
}

func (n *NodeVerifyEmail) OutputData(ctx context.Context, deps *workflow.Dependencies, w *workflow.Workflow) (interface{}, error) {
	_, resetDuration, err := deps.RateLimiter.CheckToken(n.bucket(deps))
	if err != nil {
		return nil, err
	}

	now := deps.Clock.NowUTC()
	canResendAt := now.Add(resetDuration)

	type NodeVerifyEmailOutput struct {
		Email       string    `json:"email"`
		CodeLength  int       `json:"code_length"`
		CanResendAt time.Time `json:"can_resend_at"`
	}

	return NodeVerifyEmailOutput{
		Email:       n.Email,
		CodeLength:  n.CodeLength,
		CanResendAt: canResendAt,
	}, nil
}

func (n *NodeVerifyEmail) bucket(deps *workflow.Dependencies) ratelimit.Bucket {
	return AntiSpamEmailOTPCodeBucket(deps.Config.Messaging.Email, n.Email)
}

func (n *NodeVerifyEmail) sendCode(deps *workflow.Dependencies, w *workflow.Workflow) error {
	err := deps.RateLimiter.TakeToken(n.bucket(deps))
	if err != nil {
		return err
	}

	// FIXME: web session ID?
	code, err := deps.OTPCodes.GenerateCode(n.Email, otp.OTPModeCode, string(deps.Config.ID), "")
	if err != nil {
		return err
	}
	n.CodeLength = len(code.Code)

	err = deps.OOBCodeSender.SendCode(model.AuthenticatorOOBChannelEmail, n.Email, code.Code, otp.MessageTypeVerification, otp.OTPModeCode)
	if err != nil {
		return err
	}

	return nil

}
