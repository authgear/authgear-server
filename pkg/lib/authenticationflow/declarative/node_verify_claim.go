package declarative

import (
	"context"
	"fmt"

	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	"github.com/authgear/authgear-server/pkg/api"
	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/api/model"
	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/authn"
	"github.com/authgear/authgear-server/pkg/lib/authn/otp"
	"github.com/authgear/authgear-server/pkg/lib/feature/verification"
	"github.com/authgear/authgear-server/pkg/lib/ratelimit"
	"github.com/authgear/authgear-server/pkg/lib/translation"
	"github.com/authgear/authgear-server/pkg/util/errorutil"
)

func init() {
	authflow.RegisterNode(&NodeVerifyClaim{})
}

type NodeVerifyClaim struct {
	JSONPointer          jsonpointer.T                 `json:"json_pointer,omitempty"`
	UserID               string                        `json:"user_id,omitempty"`
	Purpose              otp.Purpose                   `json:"purpose,omitempty"`
	MessageType          translation.MessageType       `json:"message_type,omitempty"`
	Form                 otp.Form                      `json:"form,omitempty"`
	ClaimName            model.ClaimName               `json:"claim_name,omitempty"`
	ClaimValue           string                        `json:"claim_value,omitempty"`
	Channel              model.AuthenticatorOOBChannel `json:"channel,omitempty"`
	WebsocketChannelName string                        `json:"websocket_channel_name,omitempty"`
}

func NewNodeVerifyClaim(ctx context.Context, deps *authflow.Dependencies, n *NodeVerifyClaim) (*authflow.NodeWithDelayedOneTimeFunction, error) {
	n.WebsocketChannelName = authflow.NewWebsocketChannelName()

	kind := n.otpKind(deps)
	simpleNode := authflow.NewNodeSimple(n)
	code, err := n.GenerateCode(ctx, deps)
	if ratelimit.IsRateLimitErrorWithBucketName(err, kind.RateLimitTriggerCooldown(n.ClaimValue).Name) {
		// Ignore trigger cooldown rate limit error; continue the flow
		code = ""
	} else if err != nil {
		return nil, err
	}

	return &authflow.NodeWithDelayedOneTimeFunction{
		Node: simpleNode,
		DelayedOneTimeFunction: func(ctx context.Context, deps *authflow.Dependencies) error {
			if code != "" {
				err := n.SendCode(ctx, deps, code)
				if err != nil {
					return err
				}
			}

			return nil
		},
	}, nil
}

var _ authflow.NodeSimple = &NodeVerifyClaim{}
var _ authflow.InputReactor = &NodeVerifyClaim{}
var _ authflow.DataOutputer = &NodeVerifyClaim{}

func (n *NodeVerifyClaim) Kind() string {
	return "NodeVerifyClaim"
}

func (n *NodeVerifyClaim) CanReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) (authflow.InputSchema, error) {
	flowRootObject, err := findNearestFlowObjectInFlow(deps, flows, n)
	if err != nil {
		return nil, err
	}

	return &InputSchemaNodeVerifyClaim{
		JSONPointer:    n.JSONPointer,
		FlowRootObject: flowRootObject,
		OTPForm:        n.Form,
	}, nil
}

func (n *NodeVerifyClaim) ReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows, input authflow.Input) (authflow.ReactToResult, error) {
	var inputNodeVerifyClaim inputNodeVerifyClaim
	if !authflow.AsInput(input, &inputNodeVerifyClaim) {
		return nil, authflow.ErrIncompatibleInput
	}

	switch {
	case inputNodeVerifyClaim.IsCode():
		code := inputNodeVerifyClaim.GetCode()

		err := deps.OTPCodes.VerifyOTP(ctx,
			n.otpKind(deps),
			n.ClaimValue,
			code,
			&otp.VerifyOptions{UserID: n.UserID},
		)

		if apierrors.IsKind(err, otp.InvalidOTPCode) {
			return nil, n.invalidOTPCodeError()
		} else if err != nil {
			return nil, err
		}

		verifiedClaim := deps.Verification.NewVerifiedClaim(ctx,
			n.UserID,
			string(n.ClaimName),
			n.ClaimValue,
		)
		verifiedClaim.SetVerifiedByChannel(n.Channel)
		return authflow.NewNodeSimple(&NodeDoMarkClaimVerified{
			Claim: verifiedClaim,
		}), nil
	case inputNodeVerifyClaim.IsCheck():
		emptyCode := ""

		err := deps.OTPCodes.VerifyOTP(ctx,
			n.otpKind(deps),
			n.ClaimValue,
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

		verifiedClaim := deps.Verification.NewVerifiedClaim(ctx,
			n.UserID,
			string(n.ClaimName),
			n.ClaimValue,
		)
		verifiedClaim.SetVerifiedByChannel(n.Channel)
		return authflow.NewNodeSimple(&NodeDoMarkClaimVerified{
			Claim: verifiedClaim,
		}), nil
	case inputNodeVerifyClaim.IsResend():
		code, err := n.GenerateCode(ctx, deps)
		if err != nil {
			return nil, err
		}

		newSimpleNode := authflow.NewNodeSimple(n)
		return &authflow.NodeWithDelayedOneTimeFunction{
			Node: newSimpleNode,
			DelayedOneTimeFunction: func(ctx context.Context, deps *authflow.Dependencies) error {
				return n.SendCode(ctx, deps, code)
			},
		}, authflow.ErrReplaceNode
	default:
		return nil, authflow.ErrIncompatibleInput
	}
}

func (n *NodeVerifyClaim) OutputData(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) (authflow.Data, error) {
	state, err := deps.OTPCodes.InspectState(ctx, n.otpKind(deps), n.ClaimValue)
	if err != nil {
		return nil, err
	}

	websocketURL := ""
	switch n.Form {
	case otp.FormLink:
		websocketURL, err = authflow.WebsocketURL(string(deps.HTTPOrigin), n.WebsocketChannelName)
		if err != nil {
			return nil, err
		}
	}

	return NewVerifyOOBOTPData(VerifyOOBOTPData{
		Channel:                        n.Channel,
		OTPForm:                        n.Form,
		WebsocketURL:                   websocketURL,
		MaskedClaimValue:               getMaskedOTPTarget(n.ClaimName, n.ClaimValue),
		CodeLength:                     n.Form.CodeLength(),
		CanResendAt:                    state.CanResendAt,
		CanCheck:                       state.SubmittedCode != "",
		FailedAttemptRateLimitExceeded: state.TooManyAttempts,
		DeliveryStatus:                 state.DeliveryStatus,
	}), nil
}

func (n *NodeVerifyClaim) otpKind(deps *authflow.Dependencies) otp.Kind {
	switch n.Purpose {
	case otp.PurposeVerification:
		return otp.KindVerification(deps.Config, n.Channel)
	case otp.PurposeOOBOTP:
		return otp.KindOOBOTPWithForm(deps.Config, n.Channel, n.Form)
	default:
		panic(fmt.Errorf("unexpected otp purpose: %v", n.Purpose))
	}
}

func (n *NodeVerifyClaim) invalidOTPCodeError() error {
	switch n.Purpose {
	case otp.PurposeVerification:
		return verification.ErrInvalidVerificationCode
	case otp.PurposeOOBOTP:
		var authenticationType authn.AuthenticationType
		switch n.Channel {
		case model.AuthenticatorOOBChannelEmail:
			authenticationType = authn.AuthenticationTypeOOBOTPEmail
		case model.AuthenticatorOOBChannelSMS:
			authenticationType = authn.AuthenticationTypeOOBOTPSMS
		case model.AuthenticatorOOBChannelWhatsapp:
			authenticationType = authn.AuthenticationTypeOOBOTPSMS
		default:
			panic(fmt.Errorf("unexpected channel: %v", n.Channel))
		}

		return errorutil.WithDetails(api.ErrInvalidCredentials, errorutil.Details{
			"AuthenticationType": apierrors.APIErrorDetail.Value(authenticationType),
		})
	default:
		panic(fmt.Errorf("unexpected otp purpose: %v", n.Purpose))
	}
}

func (n *NodeVerifyClaim) GenerateCode(ctx context.Context, deps *authflow.Dependencies) (string, error) {
	code, err := deps.OTPCodes.GenerateOTP(ctx,
		n.otpKind(deps),
		n.ClaimValue,
		n.Form,
		&otp.GenerateOptions{
			UserID:                                 n.UserID,
			AuthenticationFlowWebsocketChannelName: n.WebsocketChannelName,
		},
	)
	if err != nil {
		return "", err
	}
	return code, nil
}

func (n *NodeVerifyClaim) SendCode(ctx context.Context, deps *authflow.Dependencies, code string) error {
	// Here is a bit tricky.
	// Normally we should use the given message type to send a message.
	// However, if the channel is whatsapp, we use the specialized otp.MessageTypeWhatsappCode.
	// It is because otp.MessageTypeWhatsappCode will send a Whatsapp authentication message.
	// which is optimized for delivering a authentication code to the end-user.
	// See https://developers.facebook.com/docs/whatsapp/business-management-api/authentication-templates/
	typ := n.MessageType
	if n.Channel == model.AuthenticatorOOBChannelWhatsapp {
		typ = translation.MessageTypeWhatsappCode
	}

	err := deps.OTPSender.Send(
		ctx,
		otp.SendOptions{
			Channel: n.Channel,
			Target:  n.ClaimValue,
			Form:    n.Form,
			Kind:    n.otpKind(deps),
			Type:    typ,
			OTP:     code,
		},
	)
	if err != nil {
		return err
	}

	return nil
}
