package nodes

import (
	"errors"

	"github.com/authgear/authgear-server/pkg/lib/authn"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator/oob"
	"github.com/authgear/authgear-server/pkg/lib/authn/otp"
	"github.com/authgear/authgear-server/pkg/lib/interaction"
)

func cloneAuthenticator(info *authenticator.Info) *authenticator.Info {
	newInfo := *info
	newInfo.Claims = map[string]interface{}{}
	for k, v := range info.Claims {
		newInfo.Claims[k] = v
	}
	return &newInfo
}

func filterAuthenticators(ais []*authenticator.Info, filters ...authenticator.Filter) (out []*authenticator.Info) {
	for _, a := range ais {
		keep := true
		for _, f := range filters {
			if !f.Keep(a) {
				keep = false
				break
			}
		}
		if keep {
			out = append(out, a)
		}
	}
	return
}

func sendOOBCode(
	ctx *interaction.Context,
	stage interaction.AuthenticationStage,
	isAuthenticating bool,
	authenticatorInfo *authenticator.Info,
) (*otp.CodeSendResult, error) {
	// TODO(interaction): handle rate limits

	channel := authn.AuthenticatorOOBChannel(authenticatorInfo.Claims[authenticator.AuthenticatorClaimOOBOTPChannelType].(string))

	var messageType otp.MessageType
	switch stage {
	case interaction.AuthenticationStagePrimary:
		if isAuthenticating {
			messageType = otp.MessageTypeAuthenticatePrimaryOOB
		} else {
			messageType = otp.MessageTypeSetupPrimaryOOB
		}
	case interaction.AuthenticationStageSecondary:
		if isAuthenticating {
			messageType = otp.MessageTypeAuthenticateSecondaryOOB
		} else {
			messageType = otp.MessageTypeSetupSecondaryOOB
		}
	default:
		panic("interaction: unknown authentication stage: " + stage)
	}

	var target string
	switch channel {
	case authn.AuthenticatorOOBChannelSMS:
		target = authenticatorInfo.Claims[authenticator.AuthenticatorClaimOOBOTPPhone].(string)
	case authn.AuthenticatorOOBChannelEmail:
		target = authenticatorInfo.Claims[authenticator.AuthenticatorClaimOOBOTPEmail].(string)
	}

	code, err := ctx.OOBAuthenticators.GetCode(authenticatorInfo.ID)
	if errors.Is(err, oob.ErrCodeNotFound) {
		code = nil
	} else if err != nil {
		return nil, err
	}

	if code == nil || ctx.Clock.NowUTC().After(code.ExpireAt) {
		code, err = ctx.OOBAuthenticators.CreateCode(authenticatorInfo.ID)
		if err != nil {
			return nil, err
		}
	}

	return ctx.OOBCodeSender.SendCode(channel, target, code.Code, messageType)
}

func stageToAuthenticatorKind(stage interaction.AuthenticationStage) authenticator.Kind {
	switch stage {
	case interaction.AuthenticationStagePrimary:
		return authenticator.KindPrimary
	case interaction.AuthenticationStageSecondary:
		return authenticator.KindSecondary
	default:
		panic("interaction: unknown stage: " + stage)
	}
}
