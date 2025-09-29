package nodes

import (
	"context"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/authn/otp"
	"github.com/authgear/authgear-server/pkg/lib/feature"
	"github.com/authgear/authgear-server/pkg/lib/interaction"
	"github.com/authgear/authgear-server/pkg/lib/ratelimit"
	"github.com/authgear/authgear-server/pkg/lib/translation"
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

func (p *SendOOBCode) Do(goCtx context.Context) (*SendOOBCodeResult, error) {
	var messageType translation.MessageType
	switch p.Stage {
	case authn.AuthenticationStagePrimary:
		if p.IsAuthenticating {
			messageType = translation.MessageTypeAuthenticatePrimaryOOB
		} else {
			messageType = translation.MessageTypeSetupPrimaryOOB
		}
	case authn.AuthenticationStageSecondary:
		if p.IsAuthenticating {
			messageType = translation.MessageTypeAuthenticateSecondaryOOB
		} else {
			messageType = translation.MessageTypeSetupSecondaryOOB
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

	kind := otp.KindOOBOTPWithForm(p.Context.Config, channel, p.OTPForm)
	code, err := p.Context.OTPCodeService.GenerateOTP(
		goCtx,
		kind,
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

	err = p.Context.OTPSender.Send(
		goCtx,
		otp.SendOptions{
			Channel: channel,
			Target:  p.AuthenticatorInfo.OOBOTP.ToTarget(),
			Form:    p.OTPForm,
			Kind:    kind,
			Type:    messageType,
			OTP:     code,
		},
	)
	if p.IgnoreRatelimitError && apierrors.IsKind(err, ratelimit.RateLimited) {
		// Ignore the rate limit error and do NOT send the code.
		return result, nil
	} else if err != nil {
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
