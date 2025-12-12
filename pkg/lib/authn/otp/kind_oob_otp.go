package otp

import (
	"fmt"
	"time"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/ratelimit"
)

const PurposeOOBOTP Purpose = "oob-otp"

type kindOOBOTP struct {
	config  *config.AppConfig
	channel model.AuthenticatorOOBChannel
	purpose Purpose
	form    Form
}

func KindOOBOTPCode(config *config.AppConfig, channel model.AuthenticatorOOBChannel) Kind {
	return kindOOBOTP{config: config, channel: channel, purpose: PurposeOOBOTP, form: FormCode}
}

func KindOOBOTPLink(config *config.AppConfig, channel model.AuthenticatorOOBChannel) Kind {
	return kindOOBOTP{config: config, channel: channel, purpose: PurposeOOBOTP, form: FormLink}
}

func KindOOBOTPWithForm(config *config.AppConfig, channel model.AuthenticatorOOBChannel, form Form) Kind {
	return kindOOBOTP{config: config, channel: channel, purpose: PurposeOOBOTP, form: form}
}

var _ DeprecatedKindFactory = KindOOBOTPCode
var _ DeprecatedKindFactory = KindOOBOTPLink

func (k kindOOBOTP) Purpose() Purpose {
	return k.purpose
}

func (k kindOOBOTP) ValidPeriod() time.Duration {
	switch k.form {
	case FormCode:
		return selectByChannel(k.channel,
			k.config.Authenticator.OOB.Email.ValidPeriods.Code,
			k.config.Authenticator.OOB.SMS.ValidPeriods.Code,
			k.config.Authenticator.OOB.SMS.ValidPeriods.Code,
		).Duration()
	case FormLink:
		return selectByChannel(k.channel,
			k.config.Authenticator.OOB.Email.ValidPeriods.Link,
			k.config.Authenticator.OOB.SMS.ValidPeriods.Link,
			k.config.Authenticator.OOB.SMS.ValidPeriods.Link,
		).Duration()
	}
	panic("unknown authenticator oob otp form")
}

func (k kindOOBOTP) RateLimitTrigger(
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
		return ratelimit.RateLimitGroupAuthenticationOOBOTPEmailTrigger.ResolveBucketSpecs(k.config, featureConfig, envConfig, opts)
	case model.AuthenticatorOOBChannelSMS:
		fallthrough
	case model.AuthenticatorOOBChannelWhatsapp:
		return ratelimit.RateLimitGroupAuthenticationOOBOTPSMSTrigger.ResolveBucketSpecs(k.config, featureConfig, envConfig, opts)
	}
	panic(fmt.Errorf("invalid channel: %v", k.channel))
}

func (k kindOOBOTP) RateLimitTriggerCooldown(target string) ratelimit.BucketSpec {
	return ratelimit.NewCooldownSpec(
		selectByChannel(k.channel,
			ratelimit.OOBOTPCooldownEmail,
			ratelimit.OOBOTPCooldownSMS,
			ratelimit.OOBOTPCooldownWhatsapp,
		),
		selectByChannel(k.channel,
			k.config.Authentication.RateLimits.OOBOTP.Email.TriggerCooldown,
			k.config.Authentication.RateLimits.OOBOTP.SMS.TriggerCooldown,
			k.config.Authentication.RateLimits.OOBOTP.SMS.TriggerCooldown,
		).Duration(),
		string(k.purpose), target,
	)
}

func (k kindOOBOTP) RateLimitValidate(
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
		return ratelimit.RateLimitGroupAuthenticationOOBOTPEmailValidate.ResolveBucketSpecs(k.config, featureConfig, envConfig, opts)
	case model.AuthenticatorOOBChannelSMS:
		fallthrough
	case model.AuthenticatorOOBChannelWhatsapp:
		return ratelimit.RateLimitGroupAuthenticationOOBOTPSMSValidate.ResolveBucketSpecs(k.config, featureConfig, envConfig, opts)
	}
	panic(fmt.Errorf("invalid channel: %v", k.channel))
}

func (k kindOOBOTP) RevocationMaxFailedAttempts() int {
	return selectByChannel(k.channel,
		k.config.Authentication.RateLimits.OOBOTP.Email.MaxFailedAttemptsRevokeOTP,
		k.config.Authentication.RateLimits.OOBOTP.SMS.MaxFailedAttemptsRevokeOTP,
		k.config.Authentication.RateLimits.OOBOTP.SMS.MaxFailedAttemptsRevokeOTP,
	)
}
