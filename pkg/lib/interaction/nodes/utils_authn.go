package nodes

import (
	"github.com/authgear/authgear-server/pkg/lib/authn"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/authn/otp"
	"github.com/authgear/authgear-server/pkg/lib/interaction"
)

func cloneAuthenticator(info *authenticator.Info) *authenticator.Info {
	newInfo := *info
	newInfo.Props = map[string]interface{}{}
	for k, v := range info.Props {
		newInfo.Props[k] = v
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
	secret string,
) (*otp.CodeSendResult, error) {
	// TODO(interaction): handle rate limits

	channel := authn.AuthenticatorOOBChannel(authenticatorInfo.Props[authenticator.AuthenticatorPropOOBOTPChannelType].(string))

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
		target = authenticatorInfo.Props[authenticator.AuthenticatorPropOOBOTPPhone].(string)
	case authn.AuthenticatorOOBChannelEmail:
		target = authenticatorInfo.Props[authenticator.AuthenticatorPropOOBOTPEmail].(string)
	}

	code := ctx.OOBAuthenticators.GenerateCode(secret, channel)
	return ctx.OOBAuthenticators.SendCode(channel, target, code, messageType)
}

func stageToAuthenticatorTag(stage interaction.AuthenticationStage) []string {
	switch stage {
	case interaction.AuthenticationStagePrimary:
		return []string{authenticator.TagPrimaryAuthenticator}
	case interaction.AuthenticationStageSecondary:
		return []string{authenticator.TagSecondaryAuthenticator}
	default:
		panic("interaction: unknown stage: " + stage)
	}
}
