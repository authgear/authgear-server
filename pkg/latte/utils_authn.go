package latte

import (
	"context"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/authn/otp"
	"github.com/authgear/authgear-server/pkg/lib/feature"
	"github.com/authgear/authgear-server/pkg/lib/ratelimit"
	"github.com/authgear/authgear-server/pkg/lib/translation"
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

func (p *SendOOBCode) Do(ctx context.Context) error {
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

	kind := otp.KindOOBOTPCode(p.Deps.Config, channel)

	code, err := p.Deps.OTPCodes.GenerateOTP(
		ctx,
		kind,
		target,
		p.OTPForm,
		&otp.GenerateOptions{
			UserID:     p.AuthenticatorInfo.UserID,
			WorkflowID: p.WorkflowID,
		})
	if !p.IsResend && ratelimit.IsRateLimitErrorWithBucketName(err, kind.RateLimitTriggerCooldown(target).Name) {
		// Ignore trigger cooldown rate limit error for initial sending, and do NOT send the code.
		return nil
	} else if err != nil {
		return err
	}

	err = p.Deps.OTPSender.Send(
		ctx,
		otp.SendOptions{
			Channel: channel,
			Target:  target,
			Form:    p.OTPForm,
			Type:    messageType,
			OTP:     code,
		},
	)
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
