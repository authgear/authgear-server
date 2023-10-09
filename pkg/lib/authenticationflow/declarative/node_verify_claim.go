package declarative

import (
	"context"
	"fmt"
	"time"

	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	"github.com/authgear/authgear-server/pkg/api"
	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/api/model"
	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/authn/otp"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/feature/verification"
	"github.com/authgear/authgear-server/pkg/lib/infra/mail"
	"github.com/authgear/authgear-server/pkg/util/phone"
)

func init() {
	authflow.RegisterNode(&NodeVerifyClaim{})
}

type NodeVerifyClaimData struct {
	Channel                        model.AuthenticatorOOBChannel `json:"channel,omitempty"`
	OTPForm                        otp.Form                      `json:"otp_form,omitempty"`
	MaskedClaimValue               string                        `json:"masked_claim_value,omitempty"`
	CodeLength                     int                           `json:"code_length,omitempty"`
	CanResendAt                    time.Time                     `json:"can_resend_at,omitempty"`
	FailedAttemptRateLimitExceeded bool                          `json:"failed_attempt_rate_limit_exceeded"`
}

var _ authflow.Data = &NodeVerifyClaimData{}

func (m NodeVerifyClaimData) Data() {}

type NodeVerifyClaim struct {
	JSONPointer jsonpointer.T                 `json:"json_pointer,omitempty"`
	UserID      string                        `json:"user_id,omitempty"`
	Purpose     otp.Purpose                   `json:"purpose,omitempty"`
	MessageType otp.MessageType               `json:"message_type,omitempty"`
	ClaimName   model.ClaimName               `json:"claim_name,omitempty"`
	ClaimValue  string                        `json:"claim_value,omitempty"`
	Channel     model.AuthenticatorOOBChannel `json:"channel,omitempty"`
}

var _ authflow.NodeSimple = &NodeVerifyClaim{}
var _ authflow.InputReactor = &NodeVerifyClaim{}
var _ authflow.DataOutputer = &NodeVerifyClaim{}

func (n *NodeVerifyClaim) Kind() string {
	return "NodeVerifyClaim"
}

func (n *NodeVerifyClaim) CanReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) (authflow.InputSchema, error) {
	return &InputSchemaNodeVerifyClaim{
		JSONPointer: n.JSONPointer,
		OTPForm:     n.otpForm(deps),
	}, nil
}

func (n *NodeVerifyClaim) ReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows, input authflow.Input) (*authflow.Node, error) {
	var inputNodeVerifyClaim inputNodeVerifyClaim
	if !authflow.AsInput(input, &inputNodeVerifyClaim) {
		return nil, authflow.ErrIncompatibleInput
	}

	switch {
	case inputNodeVerifyClaim.IsCode():
		code := inputNodeVerifyClaim.GetCode()

		err := deps.OTPCodes.VerifyOTP(
			n.otpKind(deps),
			n.otpTarget(),
			code,
			&otp.VerifyOptions{UserID: n.UserID},
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
		return authflow.NewNodeSimple(&NodeDoMarkClaimVerified{
			Claim: verifiedClaim,
		}), nil
	case inputNodeVerifyClaim.IsCheck():
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
		return authflow.NewNodeSimple(&NodeDoMarkClaimVerified{
			Claim: verifiedClaim,
		}), nil
	case inputNodeVerifyClaim.IsResend():
		err := n.SendCode(ctx, deps)
		if err != nil {
			return nil, err
		}
		return authflow.NewNodeSimple(n), authflow.ErrUpdateNode
	default:
		return nil, authflow.ErrIncompatibleInput
	}
}

func (n *NodeVerifyClaim) OutputData(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) (authflow.Data, error) {
	state, err := deps.OTPCodes.InspectState(n.otpKind(deps), n.otpTarget())
	if err != nil {
		return nil, err
	}

	otpForm := n.otpForm(deps)

	return NodeVerifyClaimData{
		Channel:                        n.Channel,
		OTPForm:                        otpForm,
		MaskedClaimValue:               n.maskedOTPTarget(),
		CodeLength:                     otpForm.CodeLength(),
		CanResendAt:                    state.CanResendAt,
		FailedAttemptRateLimitExceeded: state.TooManyAttempts,
	}, nil
}

func (n *NodeVerifyClaim) otpKind(deps *authflow.Dependencies) otp.Kind {
	switch n.Purpose {
	case otp.PurposeVerification:
		return otp.KindVerification(deps.Config, n.Channel)
	case otp.PurposeOOBOTP:
		return otp.KindOOBOTP(deps.Config, n.Channel)
	default:
		panic(fmt.Errorf("unexpected otp purpose: %v", n.Purpose))
	}
}

func (n *NodeVerifyClaim) otpForm(deps *authflow.Dependencies) otp.Form {
	if n.Purpose == otp.PurposeOOBOTP &&
		n.Channel == model.AuthenticatorOOBChannelEmail &&
		deps.Config.Authenticator.OOB.Email.EmailOTPMode == config.AuthenticatorEmailOTPModeLoginLinkOnly {
		return otp.FormLink
	}
	return otp.FormCode
}

func (n *NodeVerifyClaim) invalidOTPCodeError() error {
	// FIXME(authflow): Add AuthenticationType to error info.
	switch n.Purpose {
	case otp.PurposeVerification:
		return verification.ErrInvalidVerificationCode
	case otp.PurposeOOBOTP:
		return api.ErrInvalidCredentials
	default:
		panic(fmt.Errorf("unexpected otp purpose: %v", n.Purpose))
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
		panic(fmt.Errorf("unexpected claim name: %v", n.ClaimName))
	}
}

func (n *NodeVerifyClaim) SendCode(ctx context.Context, deps *authflow.Dependencies) error {
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
			UserID:               n.UserID,
			AuthenticationFlowID: authflow.GetFlowID(ctx),
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
