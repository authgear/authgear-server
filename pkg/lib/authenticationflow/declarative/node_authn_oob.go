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
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/authn/otp"
	"github.com/authgear/authgear-server/pkg/lib/facade"
	"github.com/authgear/authgear-server/pkg/lib/feature/verification"
	"github.com/authgear/authgear-server/pkg/lib/ratelimit"
	"github.com/authgear/authgear-server/pkg/lib/translation"
	"github.com/authgear/authgear-server/pkg/util/errorutil"
)

func init() {
	authflow.RegisterNode(&NodeAuthenticationOOB{})
}

type NodeAuthenticationOOB struct {
	JSONPointer          jsonpointer.T                          `json:"json_pointer,omitempty"`
	UserID               string                                 `json:"user_id,omitempty"`
	Purpose              otp.Purpose                            `json:"purpose,omitempty"`
	Form                 otp.Form                               `json:"form,omitempty"`
	Info                 *authenticator.Info                    `json:"info,omitempty"`
	Channel              model.AuthenticatorOOBChannel          `json:"channel,omitempty"`
	WebsocketChannelName string                                 `json:"websocket_channel_name,omitempty"`
	Authentication       model.AuthenticationFlowAuthentication `json:"authentication,omitempty"`
}

func NewNodeAuthenticationOOB(n *NodeAuthenticationOOB) *authflow.NodeWithDelayedOneTimeFunction {
	n.WebsocketChannelName = authflow.NewWebsocketChannelName()
	simpleNode := authflow.NewNodeSimple(n)

	return &authflow.NodeWithDelayedOneTimeFunction{
		Node: simpleNode,
		DelayedOneTimeFunction: func(ctx context.Context, deps *authflow.Dependencies) error {
			kind := n.otpKind(deps)
			err := n.SendCode(ctx, deps)
			_, claimValue := n.Info.OOBOTP.ToClaimPair()
			if ratelimit.IsRateLimitErrorWithBucketName(err, kind.RateLimitTriggerCooldown(claimValue).Name) {
				// Ignore trigger cooldown rate limit error; continue the flow
			} else if err != nil {
				return err
			}
			return nil
		},
	}
}

var _ authflow.NodeSimple = &NodeAuthenticationOOB{}
var _ authflow.InputReactor = &NodeAuthenticationOOB{}
var _ authflow.DataOutputer = &NodeAuthenticationOOB{}

func (n *NodeAuthenticationOOB) Kind() string {
	return "NodeAuthenticationOOB"
}

func (n *NodeAuthenticationOOB) CanReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) (authflow.InputSchema, error) {
	flowRootObject, err := findNearestFlowObjectInFlow(deps, flows, n)
	if err != nil {
		return nil, err
	}

	return &InputSchemaNodeAuthenticationOOB{
		JSONPointer:    n.JSONPointer,
		FlowRootObject: flowRootObject,
		OTPForm:        n.Form,
	}, nil
}

func (n *NodeAuthenticationOOB) ReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows, input authflow.Input) (authflow.ReactToResult, error) {
	var inputNodeAuthenticationOOB inputNodeAuthenticationOOB
	if !authflow.AsInput(input, &inputNodeAuthenticationOOB) {
		return nil, authflow.ErrIncompatibleInput
	}

	switch {
	case inputNodeAuthenticationOOB.IsCode():
		code := inputNodeAuthenticationOOB.GetCode()
		claimName, claimValue := n.Info.OOBOTP.ToClaimPair()

		authenticatorSpec := n.createAuthenticatorSpec(code)
		authenticators := []*authenticator.Info{n.Info}

		_, _, err := deps.Authenticators.VerifyOneWithSpec(ctx,
			n.UserID,
			n.Info.Type,
			authenticators,
			authenticatorSpec,
			&facade.VerifyOptions{
				AuthenticationDetails: facade.NewAuthenticationDetails(
					n.UserID,
					authn.AuthenticationStageFromAuthenticationMethod(n.Authentication),
					authn.AuthenticationType(n.Info.Type),
				),
				Form: n.Form,
			},
		)
		if apierrors.IsKind(err, otp.InvalidOTPCode) {
			return nil, n.invalidOTPCodeError()
		} else if err != nil {
			return nil, err
		}

		verifiedClaim := deps.Verification.NewVerifiedClaim(ctx,
			n.UserID,
			string(claimName),
			claimValue,
		)
		return authflow.NewNodeSimple(&NodeDoMarkClaimVerified{
			Claim: verifiedClaim,
		}), nil
	case inputNodeAuthenticationOOB.IsCheck():
		emptyCode := ""
		claimName, claimValue := n.Info.OOBOTP.ToClaimPair()

		authenticatorSpec := n.createAuthenticatorSpec(emptyCode)
		authenticators := []*authenticator.Info{n.Info}

		_, _, err := deps.Authenticators.VerifyOneWithSpec(ctx,
			n.UserID,
			n.Info.Type,
			authenticators,
			authenticatorSpec,
			&facade.VerifyOptions{
				UseSubmittedValue: true,
				AuthenticationDetails: facade.NewAuthenticationDetails(
					n.UserID,
					authn.AuthenticationStageFromAuthenticationMethod(n.Authentication),
					authn.AuthenticationType(n.Info.Type),
				),
				Form: n.Form,
			},
		)
		if apierrors.IsKind(err, otp.InvalidOTPCode) {
			return nil, n.invalidOTPCodeError()
		} else if err != nil {
			return nil, err
		}

		verifiedClaim := deps.Verification.NewVerifiedClaim(ctx,
			n.UserID,
			string(claimName),
			claimValue,
		)
		return authflow.NewNodeSimple(&NodeDoMarkClaimVerified{
			Claim: verifiedClaim,
		}), nil
	case inputNodeAuthenticationOOB.IsResend():
		newSimpleNode := authflow.NewNodeSimple(n)
		return &authflow.NodeWithDelayedOneTimeFunction{
			Node: newSimpleNode,
			DelayedOneTimeFunction: func(ctx context.Context, deps *authflow.Dependencies) error {
				return n.SendCode(ctx, deps)
			},
		}, authflow.ErrReplaceNode
	default:
		return nil, authflow.ErrIncompatibleInput
	}
}

func (n *NodeAuthenticationOOB) OutputData(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) (authflow.Data, error) {
	claimName, claimValue := n.Info.OOBOTP.ToClaimPair()
	state, err := deps.OTPCodes.InspectState(ctx, n.otpKind(deps), claimValue)
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
		MaskedClaimValue:               getMaskedOTPTarget(claimName, claimValue),
		CodeLength:                     n.Form.CodeLength(),
		CanResendAt:                    state.CanResendAt,
		CanCheck:                       state.SubmittedCode != "",
		FailedAttemptRateLimitExceeded: state.TooManyAttempts,
	}), nil
}

func (*NodeAuthenticationOOB) otpMessageType(info *authenticator.Info) translation.MessageType {
	switch info.Kind {
	case model.AuthenticatorKindPrimary:
		return translation.MessageTypeAuthenticatePrimaryOOB
	case model.AuthenticatorKindSecondary:
		return translation.MessageTypeAuthenticateSecondaryOOB
	default:
		panic(fmt.Errorf("unexpected OOB OTP authenticator kind: %v", info.Kind))
	}
}

func (n *NodeAuthenticationOOB) otpKind(deps *authflow.Dependencies) otp.Kind {
	switch n.Purpose {
	case otp.PurposeVerification:
		return otp.KindVerification(deps.Config, n.Channel)
	case otp.PurposeOOBOTP:
		return otp.KindOOBOTPWithForm(deps.Config, n.Channel, n.Form)
	default:
		panic(fmt.Errorf("unexpected otp purpose: %v", n.Purpose))
	}
}

func (n *NodeAuthenticationOOB) invalidOTPCodeError() error {
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

func (n *NodeAuthenticationOOB) SendCode(ctx context.Context, deps *authflow.Dependencies) error {
	// Here is a bit tricky.
	// Normally we should use the given message type to send a message.
	// However, if the channel is whatsapp, we use the specialized otp.MessageTypeWhatsappCode.
	// It is because otp.MessageTypeWhatsappCode will send a Whatsapp authentication message.
	// which is optimized for delivering a authentication code to the end-user.
	// See https://developers.facebook.com/docs/whatsapp/business-management-api/authentication-templates/
	typ := n.otpMessageType(n.Info)
	if n.Channel == model.AuthenticatorOOBChannelWhatsapp {
		typ = translation.MessageTypeWhatsappCode
	}
	_, claimValue := n.Info.OOBOTP.ToClaimPair()

	code, err := deps.OTPCodes.GenerateOTP(ctx,
		n.otpKind(deps),
		claimValue,
		n.Form,
		&otp.GenerateOptions{
			UserID:                                 n.UserID,
			AuthenticationFlowWebsocketChannelName: n.WebsocketChannelName,
		},
	)
	if err != nil {
		return err
	}

	err = deps.OTPSender.Send(
		ctx,
		otp.SendOptions{
			Channel: n.Channel,
			Target:  claimValue,
			Form:    n.Form,
			Type:    typ,
			OTP:     code,
		},
	)
	if err != nil {
		return err
	}

	return nil
}

func (n *NodeAuthenticationOOB) createAuthenticatorSpec(code string) *authenticator.Spec {
	spec := &authenticator.Spec{
		OOBOTP: &authenticator.OOBOTPSpec{
			Code: code,
		},
	}

	switch n.Channel {
	case model.AuthenticatorOOBChannelEmail:
		spec.Type = model.AuthenticatorTypeOOBEmail
		spec.OOBOTP.Email = n.Info.OOBOTP.ToTarget()
	case model.AuthenticatorOOBChannelSMS:
		spec.Type = model.AuthenticatorTypeOOBSMS
		spec.OOBOTP.Phone = n.Info.OOBOTP.ToTarget()
	case model.AuthenticatorOOBChannelWhatsapp:
		spec.Type = model.AuthenticatorTypeOOBSMS
		spec.OOBOTP.Phone = n.Info.OOBOTP.ToTarget()
	default:
		panic(fmt.Errorf("unexpected channel: %v", n.Channel))
	}

	return spec
}
