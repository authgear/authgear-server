package otp

import (
	"fmt"
	"time"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/ratelimit"
)

const PurposeVerification Purpose = "verification"

type kindVerification struct {
	config  *config.AppConfig
	channel model.AuthenticatorOOBChannel
}

func KindVerification(config *config.AppConfig, channel model.AuthenticatorOOBChannel) Kind {
	return kindVerification{config: config, channel: channel}
}

var _ DeprecatedKindFactory = KindVerification

func (k kindVerification) Purpose() Purpose {
	return PurposeVerification
}

func (k kindVerification) ValidPeriod() time.Duration {
	return k.config.Verification.CodeValidPeriod.Duration()
}

func (k kindVerification) RateLimitTrigger(
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
		return ratelimit.RateLimitGroupVerificationEmailTrigger.ResolveBucketSpecs(k.config, featureConfig, envConfig, opts)
	case model.AuthenticatorOOBChannelSMS:
		fallthrough
	case model.AuthenticatorOOBChannelWhatsapp:
		return ratelimit.RateLimitGroupVerificationSMSTrigger.ResolveBucketSpecs(k.config, featureConfig, envConfig, opts)
	}
	panic(fmt.Errorf("invalid channel: %v", k.channel))
}

func (k kindVerification) RateLimitTriggerCooldown(target string) ratelimit.BucketSpec {
	return ratelimit.NewCooldownSpec(
		selectByChannel(k.channel,
			ratelimit.VerificationCooldownEmail,
			ratelimit.VerificationCooldownSMS,
			ratelimit.VerificationCooldownWhatsapp,
		),
		selectByChannel(k.channel,
			k.config.Verification.RateLimits.Email.TriggerCooldown,
			k.config.Verification.RateLimits.SMS.TriggerCooldown,
			k.config.Verification.RateLimits.SMS.TriggerCooldown,
		).Duration(),
		target,
	)
}

func (k kindVerification) RateLimitValidate(
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
		return ratelimit.RateLimitGroupVerificationEmailValidate.ResolveBucketSpecs(k.config, featureConfig, envConfig, opts)
	case model.AuthenticatorOOBChannelSMS:
		fallthrough
	case model.AuthenticatorOOBChannelWhatsapp:
		return ratelimit.RateLimitGroupVerificationSMSValidate.ResolveBucketSpecs(k.config, featureConfig, envConfig, opts)
	}
	panic(fmt.Errorf("invalid channel: %v", k.channel))
}

func (k kindVerification) RevocationMaxFailedAttempts() int {
	return selectByChannel(k.channel,
		k.config.Verification.RateLimits.Email.MaxFailedAttemptsRevokeOTP,
		k.config.Verification.RateLimits.SMS.MaxFailedAttemptsRevokeOTP,
		k.config.Verification.RateLimits.SMS.MaxFailedAttemptsRevokeOTP,
	)
}
