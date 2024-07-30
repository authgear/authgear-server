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
)

type SendOOBCodeResult struct {
	Target     string
	Channel    string
	CodeLength int
}

type SendOOBCode struct {
	Context              *interaction.Context
	Stage                authn.AuthenticationStage
	IsAuthenticating     bool
	AuthenticatorInfo    *authenticator.Info
	IgnoreRatelimitError bool
	OTPForm              otp.Form
}

func (p *SendOOBCode) Do() (*SendOOBCodeResult, error) {
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

	result := &SendOOBCodeResult{
		Channel:    string(channel),
		Target:     target,
		CodeLength: p.OTPForm.CodeLength(),
	}

	msg, err := p.Context.OTPSender.Prepare(channel, p.AuthenticatorInfo.OOBOTP.ToTarget(), p.OTPForm, messageType)
	if p.IgnoreRatelimitError && apierrors.IsKind(err, ratelimit.RateLimited) {
		// Ignore the rate limit error and do NOT send the code.
		return result, nil
	} else if err != nil {
		return nil, err
	}
	defer msg.Close()

	code, err := p.Context.OTPCodeService.GenerateOTP(
		otp.KindOOBOTPWithForm(p.Context.Config, channel, p.OTPForm),
		p.AuthenticatorInfo.OOBOTP.ToTarget(),
		p.OTPForm,
		&otp.GenerateOptions{WebSessionID: p.Context.WebSessionID},
	)
	if p.IgnoreRatelimitError && apierrors.IsKind(err, ratelimit.RateLimited) {
		// Ignore the rate limit error and do NOT send the code.
		return result, nil
	} else if err != nil {
		return nil, err
	}

	err = p.Context.OTPSender.Send(msg, otp.SendOptions{OTP: code})
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
