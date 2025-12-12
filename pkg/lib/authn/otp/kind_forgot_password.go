package otp

import (
	"fmt"
	"time"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/ratelimit"
)

const PurposeForgotPassword Purpose = "forgot-password"

type kindForgotPassword struct {
	config  *config.AppConfig
	channel model.AuthenticatorOOBChannel
	form    Form
}

func KindForgotPasswordLink(
	config *config.AppConfig,
	channel model.AuthenticatorOOBChannel) Kind {
	return kindForgotPassword{config: config, channel: channel, form: FormLink}
}

func KindForgotPasswordOTP(
	config *config.AppConfig,
	channel model.AuthenticatorOOBChannel) Kind {
	return kindForgotPassword{config: config, channel: channel, form: FormCode}
}

var _ DeprecatedKindFactory = KindForgotPasswordLink
var _ DeprecatedKindFactory = KindForgotPasswordOTP

func (k kindForgotPassword) Purpose() Purpose {
	return PurposeForgotPassword
}

func (k kindForgotPassword) ValidPeriod() time.Duration {
	switch k.form {
	case FormLink:
		return k.config.ForgotPassword.ValidPeriods.Link.Duration()
	case FormCode:
		return k.config.ForgotPassword.ValidPeriods.Code.Duration()
	}
	panic("unknown forgot password otp form")
}

func (k kindForgotPassword) RateLimitTrigger(
	featureConfig *config.FeatureConfig,
	envConfig *config.RateLimitsEnvironmentConfig,
	ip string, userID string,
) []*ratelimit.BucketSpec {
	opts := &ratelimit.ResolveBucketSpecOptions{
		IPAddress: ip,
		Channel:   k.channel,
		Purpose:   string(k.Purpose()),
		UserID:    userID,
	}
	switch k.channel {
	case model.AuthenticatorOOBChannelEmail:
		return ratelimit.RateLimitGroupForgotPasswordEmailTrigger.ResolveBucketSpecs(k.config, featureConfig, envConfig, opts)
	case model.AuthenticatorOOBChannelSMS:
		fallthrough
	case model.AuthenticatorOOBChannelWhatsapp:
		return ratelimit.RateLimitGroupForgotPasswordSMSTrigger.ResolveBucketSpecs(k.config, featureConfig, envConfig, opts)
	}
	panic(fmt.Errorf("invalid channel: %v", k.channel))
}

func (k kindForgotPassword) RateLimitTriggerCooldown(target string) ratelimit.BucketSpec {
	return ratelimit.NewCooldownSpec(
		selectByChannel(
			k.channel,
			ratelimit.ForgotPasswordCooldownEmail,
			ratelimit.ForgotPasswordCooldownSMS,
			ratelimit.ForgotPasswordCooldownWhatsapp,
		),
		selectByChannel(
			k.channel,
			k.config.ForgotPassword.RateLimits.Email.TriggerCooldown.Duration(),
			k.config.ForgotPassword.RateLimits.SMS.TriggerCooldown.Duration(),
			k.config.ForgotPassword.RateLimits.SMS.TriggerCooldown.Duration(),
		),
		target,
	)
}

func (k kindForgotPassword) RateLimitValidate(
	featureConfig *config.FeatureConfig,
	envConfig *config.RateLimitsEnvironmentConfig,
	ip string, userID string,
) []*ratelimit.BucketSpec {
	opts := &ratelimit.ResolveBucketSpecOptions{
		IPAddress: ip,
		Channel:   k.channel,
		Purpose:   string(k.Purpose()),
		UserID:    userID,
	}
	switch k.channel {
	case model.AuthenticatorOOBChannelEmail:
		return ratelimit.RateLimitGroupForgotPasswordEmailValidate.ResolveBucketSpecs(k.config, featureConfig, envConfig, opts)
	case model.AuthenticatorOOBChannelSMS:
		fallthrough
	case model.AuthenticatorOOBChannelWhatsapp:
		return ratelimit.RateLimitGroupForgotPasswordSMSValidate.ResolveBucketSpecs(k.config, featureConfig, envConfig, opts)
	}
	panic(fmt.Errorf("invalid channel: %v", k.channel))
}

func (k kindForgotPassword) RevocationMaxFailedAttempts() int {
	return 0
}
