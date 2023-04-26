package otp

import (
	"time"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/ratelimit"
)

const PurposeOOBOTP Purpose = "oob-otp"

const PurposeWhatsapp Purpose = "whatsapp"

type kindOOBOTP struct {
	config  *config.AppConfig
	channel model.AuthenticatorOOBChannel
	purpose Purpose
}

func KindOOBOTP(config *config.AppConfig, channel model.AuthenticatorOOBChannel) Kind {
	return kindOOBOTP{config: config, channel: channel, purpose: PurposeOOBOTP}
}

// KindWhatsapp is for Whatsapp OTP code;
// it uses different purpose than OOB-OTP, since it is shown to user rather than received from user.
// Reusing same purpose leads to OTP for whatsapp usable for SMS OTP.
func KindWhatsapp(config *config.AppConfig) Kind {
	return kindOOBOTP{config: config, channel: model.AuthenticatorOOBChannelSMS, purpose: PurposeWhatsapp}
}

func (k kindOOBOTP) Purpose() Purpose {
	return k.purpose
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
		"OOBOTPTrigger", string(k.purpose), "ip", ip)
}

func (k kindOOBOTP) RateLimitTriggerPerUser(userID string) ratelimit.BucketSpec {
	return ratelimit.NewBucketSpec(
		selectByChannel(k.channel,
			k.config.Authentication.RateLimits.OOBOTP.Email.TriggerPerUser,
			k.config.Authentication.RateLimits.OOBOTP.SMS.TriggerPerUser,
		),
		"OOBOTPTrigger", string(k.purpose), "user", userID)
}

func (k kindOOBOTP) RateLimitTriggerCooldown(target string) ratelimit.BucketSpec {
	return ratelimit.NewCooldownSpec(
		"OOBOTPCooldown",
		selectByChannel(k.channel,
			k.config.Authentication.RateLimits.OOBOTP.Email.TriggerCooldown,
			k.config.Authentication.RateLimits.OOBOTP.SMS.TriggerCooldown,
		).Duration(),
		string(k.purpose), target,
	)
}

func (k kindOOBOTP) RateLimitValidatePerIP(ip string) ratelimit.BucketSpec {
	config := selectByChannel(k.channel,
		k.config.Authentication.RateLimits.OOBOTP.Email.ValidatePerIP,
		k.config.Authentication.RateLimits.OOBOTP.SMS.ValidatePerIP,
	)
	if config.Enabled == nil {
		config = k.config.Authentication.RateLimits.General.PerIP
	}
	return ratelimit.NewBucketSpec(config, "OOBOTPValidateCode", string(k.purpose), "ip", ip)
}

func (k kindOOBOTP) RateLimitValidatePerUserPerIP(userID string, ip string) ratelimit.BucketSpec {
	config := selectByChannel(k.channel,
		k.config.Authentication.RateLimits.OOBOTP.Email.ValidatePerUserPerIP,
		k.config.Authentication.RateLimits.OOBOTP.SMS.ValidatePerUserPerIP,
	)
	if config.Enabled == nil {
		config = k.config.Authentication.RateLimits.General.PerUserPerIP
	}
	return ratelimit.NewBucketSpec(config, "OOBOTPValidateCode", string(k.purpose), "user", userID, "ip", ip)
}

func (k kindOOBOTP) RevocationMaxFailedAttempts() int {
	return selectByChannel(k.channel,
		k.config.Authentication.RateLimits.OOBOTP.Email.MaxFailedAttemptsRevokeOTP,
		k.config.Authentication.RateLimits.OOBOTP.SMS.MaxFailedAttemptsRevokeOTP,
	)
}
