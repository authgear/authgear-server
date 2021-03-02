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

	var messageType otp.MessageType
	var oobType interaction.OOBType
	switch stage {
	case interaction.AuthenticationStagePrimary:
		if isAuthenticating {
			messageType = otp.MessageTypeAuthenticatePrimaryOOB
			oobType = interaction.OOBTypeAuthenticatePrimary
		} else {
			messageType = otp.MessageTypeSetupPrimaryOOB
			oobType = interaction.OOBTypeSetupPrimary
		}
	case interaction.AuthenticationStageSecondary:
		if isAuthenticating {
			messageType = otp.MessageTypeAuthenticateSecondaryOOB
			oobType = interaction.OOBTypeAuthenticateSecondary
		} else {
			messageType = otp.MessageTypeSetupSecondaryOOB
			oobType = interaction.OOBTypeSetupSecondary
		}
	default:
		panic("interaction: unknown authentication stage: " + stage)
	}

	var channel authn.AuthenticatorOOBChannel
	var target string
	switch authenticatorInfo.Type {
	case authn.AuthenticatorTypeOOBSMS:
		channel = authn.AuthenticatorOOBChannelSMS
		target = authenticatorInfo.Claims[authenticator.AuthenticatorClaimOOBOTPPhone].(string)
	case authn.AuthenticatorTypeOOBEmail:
		channel = authn.AuthenticatorOOBChannelEmail
		target = authenticatorInfo.Claims[authenticator.AuthenticatorClaimOOBOTPEmail].(string)
	default:
		panic("interaction: incompatible authenticator type for sending oob code: " + authenticatorInfo.Type)
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

	err = ctx.RateLimiter.TakeToken(interaction.SendOOBCodeRateLimitBucket(oobType, target))
	if err != nil {
		return nil, err
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
