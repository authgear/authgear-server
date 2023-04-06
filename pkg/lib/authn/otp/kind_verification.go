package otp

import (
	"time"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/ratelimit"
)

type kindVerification struct {
	config  *config.AppConfig
	channel model.AuthenticatorOOBChannel
}

func KindVerification(config *config.AppConfig, channel model.AuthenticatorOOBChannel) Kind {
	return kindVerification{config: config, channel: channel}
}

func (k kindVerification) Purpose() string {
	return "verification"
}

func (k kindVerification) ValidPeriod() time.Duration {
	return k.config.Verification.CodeValidPeriod.Duration()
}

func (k kindVerification) RateLimitTriggerPerIP(ip string) ratelimit.BucketSpec {
	return ratelimit.NewBucketSpec(
		selectByChannel(k.channel,
			k.config.Verification.RateLimits.Email.TriggerPerIP,
			k.config.Verification.RateLimits.SMS.TriggerPerIP,
		),
		"VerificationTrigger", "ip", ip)
}

func (k kindVerification) RateLimitTriggerPerUser(userID string) ratelimit.BucketSpec {
	return ratelimit.NewBucketSpec(
		selectByChannel(k.channel,
			k.config.Verification.RateLimits.Email.TriggerPerUser,
			k.config.Verification.RateLimits.SMS.TriggerPerUser,
		),
		"VerificationTrigger", "user", userID)
}

func (k kindVerification) RateLimitTriggerCooldown(target string) ratelimit.BucketSpec {
	return ratelimit.BucketSpec{
		Name:      "VerificationCooldown",
		Arguments: []string{target},

		Enabled: true,
		Period: selectByChannel(k.channel,
			k.config.Verification.RateLimits.Email.TriggerCooldown,
			k.config.Verification.RateLimits.SMS.TriggerCooldown,
		).Duration(),
		Burst: 1,
	}
}

func (k kindVerification) RateLimitValidatePerIP(ip string) ratelimit.BucketSpec {
	return ratelimit.NewBucketSpec(
		selectByChannel(k.channel,
			k.config.Verification.RateLimits.Email.ValidatePerIP,
			k.config.Verification.RateLimits.SMS.ValidatePerIP,
		),
		"VerificationValidateCode", "ip", ip)
}

func (k kindVerification) RateLimitValidatePerUserPerIP(userID string, ip string) ratelimit.BucketSpec {
	return ratelimit.BucketSpecDisabled
}

func (k kindVerification) RevocationMaxFailedAttempts() int {
	return selectByChannel(k.channel,
		k.config.Verification.RateLimits.Email.MaxFailedAttemptsRevokeOTP,
		k.config.Verification.RateLimits.SMS.MaxFailedAttemptsRevokeOTP,
	)
}
