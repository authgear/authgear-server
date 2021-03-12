package nodes

import (
	"errors"

	"github.com/authgear/authgear-server/pkg/lib/authn"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator/oob"
	"github.com/authgear/authgear-server/pkg/lib/authn/otp"
	"github.com/authgear/authgear-server/pkg/lib/interaction"
	"github.com/authgear/authgear-server/pkg/lib/ratelimit"
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

type SendOOBCode struct {
	Context              *interaction.Context
	Stage                interaction.AuthenticationStage
	IsAuthenticating     bool
	AuthenticatorInfo    *authenticator.Info
	IgnoreRatelimitError bool
}

func (p *SendOOBCode) Do() (*otp.CodeSendResult, error) {
	var messageType otp.MessageType
	var oobType interaction.OOBType
	switch p.Stage {
	case interaction.AuthenticationStagePrimary:
		if p.IsAuthenticating {
			messageType = otp.MessageTypeAuthenticatePrimaryOOB
			oobType = interaction.OOBTypeAuthenticatePrimary
		} else {
			messageType = otp.MessageTypeSetupPrimaryOOB
			oobType = interaction.OOBTypeSetupPrimary
		}
	case interaction.AuthenticationStageSecondary:
		if p.IsAuthenticating {
			messageType = otp.MessageTypeAuthenticateSecondaryOOB
			oobType = interaction.OOBTypeAuthenticateSecondary
		} else {
			messageType = otp.MessageTypeSetupSecondaryOOB
			oobType = interaction.OOBTypeSetupSecondary
		}
	default:
		panic("interaction: unknown authentication stage: " + p.Stage)
	}

	var channel authn.AuthenticatorOOBChannel
	var target string
	switch p.AuthenticatorInfo.Type {
	case authn.AuthenticatorTypeOOBSMS:
		channel = authn.AuthenticatorOOBChannelSMS
		target = p.AuthenticatorInfo.Claims[authenticator.AuthenticatorClaimOOBOTPPhone].(string)
	case authn.AuthenticatorTypeOOBEmail:
		channel = authn.AuthenticatorOOBChannelEmail
		target = p.AuthenticatorInfo.Claims[authenticator.AuthenticatorClaimOOBOTPEmail].(string)
	default:
		panic("interaction: incompatible authenticator type for sending oob code: " + p.AuthenticatorInfo.Type)
	}

	code, err := p.Context.OOBAuthenticators.GetCode(p.AuthenticatorInfo.ID)
	if errors.Is(err, oob.ErrCodeNotFound) {
		code = nil
	} else if err != nil {
		return nil, err
	}

	if code == nil || p.Context.Clock.NowUTC().After(code.ExpireAt) {
		code, err = p.Context.OOBAuthenticators.CreateCode(p.AuthenticatorInfo.ID)
		if err != nil {
			return nil, err
		}
	}

	result := &otp.CodeSendResult{
		Channel:    string(channel),
		Target:     target,
		CodeLength: len(code.Code),
	}

	err = p.Context.RateLimiter.TakeToken(interaction.SendOOBCodeRateLimitBucket(oobType, target))
	if p.IgnoreRatelimitError && errors.Is(err, ratelimit.ErrTooManyRequests) {
		// Ignore the rate limit error and do NOT send the code.
		return result, nil
	} else if err != nil {
		return nil, err
	}

	err = p.Context.OOBCodeSender.SendCode(channel, target, code.Code, messageType)
	if err != nil {
		return nil, err
	}

	return result, nil
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
