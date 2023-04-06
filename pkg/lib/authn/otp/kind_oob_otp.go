package otp

import (
	"time"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/ratelimit"
)

type kindOOBOTP struct {
	config  *config.AppConfig
	channel model.AuthenticatorOOBChannel
}

func KindOOBOTP(config *config.AppConfig, channel model.AuthenticatorOOBChannel) Kind {
	return kindOOBOTP{config: config, channel: channel}
}

func (k kindOOBOTP) Purpose() string {
	return "oob-otp"
}

func (k kindOOBOTP) ValidPeriod() time.Duration {
	return selectByChannel(k.channel,
		k.config.Authenticator.OOB.Email.CodeValidPeriod,
		k.config.Authenticator.OOB.SMS.CodeValidPeriod,
	).Duration()
}

func (k kindOOBOTP) RateLimitTriggerPerIP(ip string) ratelimit.BucketSpec {
	return ratelimit.NewBucketSpec(
		selectByChannel(k.channel,
			k.config.Authentication.RateLimits.OOBOTP.Email.TriggerPerIP,
			k.config.Authentication.RateLimits.OOBOTP.SMS.TriggerPerIP,
		),
		"OOBOTPTrigger", "ip", ip)
}

func (k kindOOBOTP) RateLimitTriggerPerUser(userID string) ratelimit.BucketSpec {
	return ratelimit.NewBucketSpec(
		selectByChannel(k.channel,
			k.config.Authentication.RateLimits.OOBOTP.Email.TriggerPerUser,
			k.config.Authentication.RateLimits.OOBOTP.SMS.TriggerPerUser,
		),
		"OOBOTPTrigger", "user", userID)
}

func (k kindOOBOTP) RateLimitTriggerCooldown(target string) ratelimit.BucketSpec {
	return ratelimit.BucketSpec{
		Name:      "OOBOTPCooldown",
		Arguments: []string{target},

		Enabled: true,
		Period: selectByChannel(k.channel,
			k.config.Authentication.RateLimits.OOBOTP.Email.TriggerCooldown,
			k.config.Authentication.RateLimits.OOBOTP.SMS.TriggerCooldown,
		).Duration(),
		Burst: 1,
	}
}

func (k kindOOBOTP) RateLimitValidatePerIP(ip string) ratelimit.BucketSpec {
	config := selectByChannel(k.channel,
		k.config.Authentication.RateLimits.OOBOTP.Email.ValidatePerIP,
		k.config.Authentication.RateLimits.OOBOTP.SMS.ValidatePerIP,
	)
	if config.Enabled == nil {
		config = k.config.Authentication.RateLimits.General.PerIP
	}
	return ratelimit.NewBucketSpec(config, "OOBOTPValidateCode", "ip", ip)
}

func (k kindOOBOTP) RateLimitValidatePerUserPerIP(userID string, ip string) ratelimit.BucketSpec {
	config := selectByChannel(k.channel,
		k.config.Authentication.RateLimits.OOBOTP.Email.ValidatePerUserPerIP,
		k.config.Authentication.RateLimits.OOBOTP.SMS.ValidatePerUserPerIP,
	)
	if config.Enabled == nil {
		config = k.config.Authentication.RateLimits.General.PerUserPerIP
	}
	return ratelimit.NewBucketSpec(config, "OOBOTPValidateCode", "user", userID, "ip", ip)
}

func (k kindOOBOTP) RevocationMaxFailedAttempts() int {
	return selectByChannel(k.channel,
		k.config.Authentication.RateLimits.OOBOTP.Email.MaxFailedAttemptsRevokeOTP,
		k.config.Authentication.RateLimits.OOBOTP.SMS.MaxFailedAttemptsRevokeOTP,
	)
}
