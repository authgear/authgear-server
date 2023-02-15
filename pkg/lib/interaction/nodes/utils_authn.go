package nodes

import (
	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/authn/otp"
	"github.com/authgear/authgear-server/pkg/lib/feature"
	"github.com/authgear/authgear-server/pkg/lib/interaction"
	"github.com/authgear/authgear-server/pkg/lib/ratelimit"
	"github.com/authgear/authgear-server/pkg/util/clock"
)

type SendOOBCode struct {
	Clock                clock.Clock
	Context              *interaction.Context
	Stage                authn.AuthenticationStage
	IsAuthenticating     bool
	AuthenticatorInfo    *authenticator.Info
	IgnoreRatelimitError bool
	OTPMode              otp.OTPMode
	Code                 string
}

func (p *SendOOBCode) Do() (*otp.CodeSendResult, error) {
	var messageType otp.MessageType
	switch p.Stage {
	case authn.AuthenticationStagePrimary:
		if p.IsAuthenticating {
			messageType = otp.MessageTypeAuthenticatePrimaryOOB
		} else {
			messageType = otp.MessageTypeSetupPrimaryOOB
		}
	case authn.AuthenticationStageSecondary:
		if p.IsAuthenticating {
			messageType = otp.MessageTypeAuthenticateSecondaryOOB
		} else {
			messageType = otp.MessageTypeSetupSecondaryOOB
		}
	default:
		panic("interaction: unknown authentication stage: " + p.Stage)
	}

	var channel model.AuthenticatorOOBChannel
	var target string
	switch p.AuthenticatorInfo.Type {
	case model.AuthenticatorTypeOOBSMS:
		channel = model.AuthenticatorOOBChannelSMS
		target = p.AuthenticatorInfo.OOBOTP.Phone
	case model.AuthenticatorTypeOOBEmail:
		channel = model.AuthenticatorOOBChannelEmail
		target = p.AuthenticatorInfo.OOBOTP.Email
	default:
		panic("interaction: incompatible authenticator type for sending oob code: " + p.AuthenticatorInfo.Type)
	}

	// check for feature disabled
	if p.AuthenticatorInfo.Type == model.AuthenticatorTypeOOBSMS {
		fc := p.Context.FeatureConfig
		switch p.Stage {
		case authn.AuthenticationStagePrimary:
			if fc.Identity.LoginID.Types.Phone.Disabled {
				return nil, feature.ErrFeatureDisabledSendingSMS
			}
		case authn.AuthenticationStageSecondary:
			if fc.Authentication.SecondaryAuthenticators.OOBOTPSMS.Disabled {
				return nil, feature.ErrFeatureDisabledSendingSMS
			}
		}
	}

	code, err := p.Context.OTPCodeService.GenerateCode(p.AuthenticatorInfo.OOBOTP.ToTarget(), p.OTPMode, p.Context.WebSessionID)
	if err != nil {
		return nil, err
	}

	result := &otp.CodeSendResult{
		Channel:    string(channel),
		Target:     target,
		CodeLength: len(code.Code),
		Code:       code.Code,
	}

	bucket := p.Context.AntiSpamOTPCodeBucket.MakeBucket(channel, target)
	err = p.Context.RateLimiter.TakeToken(bucket)
	if p.IgnoreRatelimitError && apierrors.IsKind(err, ratelimit.RateLimited) {
		// Ignore the rate limit error and do NOT send the code.
		return result, nil
	} else if err != nil {
		return nil, err
	}

	err = p.Context.OOBCodeSender.SendCode(channel, target, code.Code, messageType, p.OTPMode)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func stageToAuthenticatorKind(stage authn.AuthenticationStage) authenticator.Kind {
	switch stage {
	case authn.AuthenticationStagePrimary:
		return authenticator.KindPrimary
	case authn.AuthenticationStageSecondary:
		return authenticator.KindSecondary
	default:
		panic("interaction: unknown stage: " + stage)
	}
}
