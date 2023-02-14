package latte

import (
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/authn/otp"
	"github.com/authgear/authgear-server/pkg/lib/feature"
	"github.com/authgear/authgear-server/pkg/lib/workflow"
)

type SendOOBCode struct {
	Deps              *workflow.Dependencies
	Stage             authn.AuthenticationStage
	IsAuthenticating  bool
	AuthenticatorInfo *authenticator.Info
	OTPMode           otp.OTPMode
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
		fc := p.Deps.FeatureConfig
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

	bucket := p.Deps.AntiSpamOTPCodeBucket.MakeBucket(channel, target)
	err := p.Deps.RateLimiter.TakeToken(bucket)
	if err != nil {
		return nil, err
	}

	// fixme(workflow): update web session id for magic link
	code, err := p.Deps.OTPCodes.GenerateCode(p.AuthenticatorInfo.OOBOTP.ToTarget(), p.OTPMode, string(p.Deps.Config.ID), "")
	if err != nil {
		return nil, err
	}

	result := &otp.CodeSendResult{
		Channel:    string(channel),
		Target:     target,
		CodeLength: len(code.Code),
		Code:       code.Code,
	}

	err = p.Deps.OOBCodeSender.SendCode(channel, target, code.Code, messageType, p.OTPMode)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func authenticatorKindToStage(kind authenticator.Kind) authn.AuthenticationStage {
	switch kind {
	case authenticator.KindPrimary:
		return authn.AuthenticationStagePrimary
	case authenticator.KindSecondary:
		return authn.AuthenticationStageSecondary
	default:
		panic("workflow: unexpected authenticator kind: " + kind)
	}
}
