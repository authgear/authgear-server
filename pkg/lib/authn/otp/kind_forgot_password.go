package otp

import (
	"time"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/ratelimit"
)

const PurposeForgotPassword Purpose = "forgot-password"

type kindForgotPassword struct {
	config  *config.AppConfig
	channel model.AuthenticatorOOBChannel
}

func KindForgotPassword(config *config.AppConfig, channel model.AuthenticatorOOBChannel) Kind {
	return kindForgotPassword{config: config, channel: channel}
}

func (k kindForgotPassword) Purpose() Purpose {
	return PurposeForgotPassword
}

func (k kindForgotPassword) ValidPeriod() time.Duration {
	return k.config.ForgotPassword.CodeValidPeriod.Duration()
}

func (k kindForgotPassword) RateLimitTriggerPerIP(ip string) ratelimit.BucketSpec {
	return ratelimit.NewBucketSpec(
		selectByChannel(k.channel,
			k.config.ForgotPassword.RateLimits.Email.TriggerPerIP,
			k.config.ForgotPassword.RateLimits.SMS.TriggerPerIP,
		),
		"ForgotPasswordTrigger", "ip", ip)
}

func (k kindForgotPassword) RateLimitTriggerPerUser(userID string) ratelimit.BucketSpec {
	return ratelimit.BucketSpecDisabled
}

func (k kindForgotPassword) RateLimitTriggerCooldown(target string) ratelimit.BucketSpec {
	return ratelimit.BucketSpec{
		Name:      "ForgotPasswordCooldown",
		Arguments: []string{target},

		Enabled: true,
		Period: selectByChannel(k.channel,
			k.config.ForgotPassword.RateLimits.Email.TriggerCooldown,
			k.config.ForgotPassword.RateLimits.SMS.TriggerCooldown,
		).Duration(),
		Burst: 1,
	}
}

func (k kindForgotPassword) RateLimitValidatePerIP(ip string) ratelimit.BucketSpec {
	return ratelimit.NewBucketSpec(
		selectByChannel(k.channel,
			k.config.ForgotPassword.RateLimits.Email.ValidatePerIP,
			k.config.ForgotPassword.RateLimits.SMS.ValidatePerIP,
		),
		"ForgotPasswordValidateCode", "ip", ip)
}

func (k kindForgotPassword) RateLimitValidatePerUserPerIP(userID string, ip string) ratelimit.BucketSpec {
	return ratelimit.BucketSpecDisabled
}

func (k kindForgotPassword) RevocationMaxFailedAttempts() int {
	return 0
}
