package workflowconfig

import (
	"context"
	"fmt"
	"time"

	"github.com/authgear/authgear-server/pkg/api"
	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/otp"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/feature/verification"
	"github.com/authgear/authgear-server/pkg/lib/infra/mail"
	workflow "github.com/authgear/authgear-server/pkg/lib/workflow2"
	"github.com/authgear/authgear-server/pkg/util/phone"
)

func init() {
	workflow.RegisterNode(&NodeVerifyClaim{})
}

type NodeVerifyClaim struct {
	UserID      string                        `json:"user_id,omitempty"`
	Purpose     otp.Purpose                   `json:"purpose,omitempty"`
	MessageType otp.MessageType               `json:"message_type,omitempty"`
	ClaimName   model.ClaimName               `json:"claim_name,omitempty"`
	ClaimValue  string                        `json:"claim_value,omitempty"`
	Channel     model.AuthenticatorOOBChannel `json:"channel,omitempty"`
}

var _ workflow.NodeSimple = &NodeVerifyClaim{}
var _ workflow.InputReactor = &NodeVerifyClaim{}
var _ workflow.DataOutputer = &NodeVerifyClaim{}

func (n *NodeVerifyClaim) Kind() string {
	return "workflowconfig.NodeVerifyClaim"
}

func (n *NodeVerifyClaim) CanReactTo(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) ([]workflow.Input, error) {
	return []workflow.Input{
		&InputTakeOOBOTPCode{},
		&InputCheckOOBOTPCodeVerified{},
		&InputResendOOBOTPCode{},
	}, nil
}

func (n *NodeVerifyClaim) ReactTo(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows, input workflow.Input) (*workflow.Node, error) {
	var inputTakeOOBOTPCode inputTakeOOBOTPCode
	var inputCheckOOBOTPCodeVerified inputCheckOOBOTPCodeVerified
	var inputResendOOBOTPCode inputResendOOBOTPCode

	switch {
	case workflow.AsInput(input, &inputTakeOOBOTPCode):
		code := inputTakeOOBOTPCode.GetCode()

		err := deps.OTPCodes.VerifyOTP(
			n.otpKind(deps),
			n.otpTarget(),
			code,
			&otp.VerifyOptions{UserID: n.UserID},
		)

		if apierrors.IsKind(err, otp.InvalidOTPCode) {
			return nil, n.invalidOTPCodeError()
			//return nil, verification.ErrInvalidVerificationCode
		} else if err != nil {
			return nil, err
		}

		verifiedClaim := deps.Verification.NewVerifiedClaim(
			n.UserID,
			string(n.ClaimName),
			n.otpTarget(),
		)
		return workflow.NewNodeSimple(&NodeDoMarkClaimVerified{
			Claim: verifiedClaim,
		}), nil
	case workflow.AsInput(input, &inputCheckOOBOTPCodeVerified):
		emptyCode := ""

		err := deps.OTPCodes.VerifyOTP(
			n.otpKind(deps),
			n.otpTarget(),
			emptyCode,
			&otp.VerifyOptions{
				UseSubmittedCode: true,
				UserID:           n.UserID,
			},
		)

		if apierrors.IsKind(err, otp.InvalidOTPCode) {
			return nil, n.invalidOTPCodeError()
		} else if err != nil {
			return nil, err
		}

		verifiedClaim := deps.Verification.NewVerifiedClaim(
			n.UserID,
			string(n.ClaimName),
			n.otpTarget(),
		)
		return workflow.NewNodeSimple(&NodeDoMarkClaimVerified{
			Claim: verifiedClaim,
		}), nil
	case workflow.AsInput(input, &inputResendOOBOTPCode):
		err := n.SendCode(ctx, deps)
		if err != nil {
			return nil, err
		}
		return workflow.NewNodeSimple(n), workflow.ErrUpdateNode
	default:
		return nil, workflow.ErrIncompatibleInput
	}
}

func (n *NodeVerifyClaim) OutputData(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) (interface{}, error) {
	state, err := deps.OTPCodes.InspectState(n.otpKind(deps), n.otpTarget())
	if err != nil {
		return nil, err
	}

	type NodeVerifyClaimOutput struct {
		OTPForm                        otp.Form  `json:"otp_form,omitempty"`
		MaskedClaimValue               string    `json:"masked_claim_value,omitempty"`
		CodeLength                     int       `json:"code_length,omitempty"`
		CanResendAt                    time.Time `json:"can_resend_at,omitempty"`
		FailedAttemptRateLimitExceeded bool      `json:"failed_attempt_rate_limit_exceeded"`
	}

	otpForm := n.otpForm(deps)

	return NodeVerifyClaimOutput{
		OTPForm:                        otpForm,
		MaskedClaimValue:               n.maskedOTPTarget(),
		CodeLength:                     otpForm.CodeLength(),
		CanResendAt:                    state.CanResendAt,
		FailedAttemptRateLimitExceeded: state.TooManyAttempts,
	}, nil
}

func (n *NodeVerifyClaim) otpKind(deps *workflow.Dependencies) otp.Kind {
	switch n.Purpose {
	case otp.PurposeVerification:
		return otp.KindVerification(deps.Config, n.Channel)
	case otp.PurposeOOBOTP:
		return otp.KindOOBOTP(deps.Config, n.Channel)
	default:
		panic(fmt.Errorf("workflow: unexpected otp purpose: %v", n.Purpose))
	}
}

func (n *NodeVerifyClaim) otpForm(deps *workflow.Dependencies) otp.Form {
	if n.Purpose == otp.PurposeOOBOTP &&
		n.Channel == model.AuthenticatorOOBChannelEmail &&
		deps.Config.Authenticator.OOB.Email.EmailOTPMode == config.AuthenticatorEmailOTPModeLoginLinkOnly {
		return otp.FormLink
	}
	return otp.FormCode
}

func (n *NodeVerifyClaim) invalidOTPCodeError() error {
	switch n.Purpose {
	case otp.PurposeVerification:
		return verification.ErrInvalidVerificationCode
	case otp.PurposeOOBOTP:
		return api.ErrInvalidCredentials
	default:
		panic(fmt.Errorf("workflow: unexpected otp purpose: %v", n.Purpose))
	}
}

func (n *NodeVerifyClaim) otpTarget() string {
	return n.ClaimValue
}

func (n *NodeVerifyClaim) maskedOTPTarget() string {
	switch n.ClaimName {
	case model.ClaimEmail:
		return mail.MaskAddress(n.otpTarget())
	case model.ClaimPhoneNumber:
		return phone.Mask(n.otpTarget())
	default:
		panic(fmt.Errorf("workflow: unexpected claim name: %v", n.ClaimName))
	}
}

func (n *NodeVerifyClaim) SendCode(ctx context.Context, deps *workflow.Dependencies) error {
	// Here is a bit tricky.
	// Normally we should use the given message type to send a message.
	// However, if the channel is whatsapp, we use the specialized otp.MessageTypeWhatsappCode.
	// It is because otp.MessageTypeWhatsappCode will send a Whatsapp authentication message.
	// which is optimized for delivering a authentication code to the end-user.
	// See https://developers.facebook.com/docs/whatsapp/business-management-api/authentication-templates/
	typ := n.MessageType
	if n.Channel == model.AuthenticatorOOBChannelWhatsapp {
		typ = otp.MessageTypeWhatsappCode
	}

	msg, err := deps.OTPSender.Prepare(
		n.Channel,
		n.otpTarget(),
		n.otpForm(deps),
		typ,
	)
	if err != nil {
		return err
	}
	defer msg.Close()

	code, err := deps.OTPCodes.GenerateOTP(
		n.otpKind(deps),
		n.otpTarget(),
		n.otpForm(deps),
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
