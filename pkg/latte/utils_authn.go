package latte

import (
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/authn/otp"
	"github.com/authgear/authgear-server/pkg/lib/feature"
	"github.com/authgear/authgear-server/pkg/lib/ratelimit"
	"github.com/authgear/authgear-server/pkg/lib/workflow"
)

type SendOOBCode struct {
	WorkflowID        string
	Deps              *workflow.Dependencies
	Stage             authn.AuthenticationStage
	IsAuthenticating  bool
	AuthenticatorInfo *authenticator.Info
	OTPForm           otp.Form
	IsResend          bool
}

func (p *SendOOBCode) Do() error {
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
				return feature.ErrFeatureDisabledSendingSMS
			}
		case authn.AuthenticationStageSecondary:
			if fc.Authentication.SecondaryAuthenticators.OOBOTPSMS.Disabled {
				return feature.ErrFeatureDisabledSendingSMS
			}
		}
	}

	// Should check if we can send code to the target first before taking token
	// from the AntiSpamOTPCodeBucket (resend cooldown)
	// It may be blocked due to exceeding the per target or per ip rate limit,
	// and this error should be returned
	var err error
	err = p.Deps.OOBCodeSender.CanSendCode(channel, target)
	if err != nil {
		return err
	}

	kind := otp.KindOOBOTP(p.Deps.Config, channel, p.OTPForm)
	code, err := p.Deps.OTPCodes.GenerateOTP(
		kind,
		target,
		&otp.GenerateOptions{
			UserID:     p.AuthenticatorInfo.UserID,
			WorkflowID: p.WorkflowID,
		})
	if !p.IsResend && ratelimit.IsRateLimitErrorWithBucketName(err, kind.RateLimitTriggerCooldown(target).Name) {
		// Ignore trigger cooldown rate limit error for initial sending
	} else if err != nil {
		return err
	}

	// FIXME: mode -> form
	var mode otp.OTPMode
	switch p.OTPForm {
	case otp.FormCode:
		mode = otp.OTPModeCode
	case otp.FormLink:
		mode = otp.OTPModeLoginLink
	}
	err = p.Deps.OOBCodeSender.SendCode(channel, target, code, messageType, mode)
	if err != nil {
		return err
	}

	return nil
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
