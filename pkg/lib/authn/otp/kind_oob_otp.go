package otp

import (
	"time"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/ratelimit"
)

const PurposeOOBOTP Purpose = "oob-otp"

const (
	OOBOTPTriggerEmailPerIP         ratelimit.BucketName = "OOBOTPTriggerEmailPerIP"
	OOBOTPTriggerSMSPerIP           ratelimit.BucketName = "OOBOTPTriggerSMSPerIP"
	OOBOTPTriggerEmailPerUser       ratelimit.BucketName = "OOBOTPTriggerEmailPerUser"
	OOBOTPTriggerSMSPerUser         ratelimit.BucketName = "OOBOTPTriggerSMSPerUser"
	OOBOTPCooldownEmail             ratelimit.BucketName = "OOBOTPCooldownEmail"
	OOBOTPCooldownSMS               ratelimit.BucketName = "OOBOTPCooldownSMS"
	OOBOTPValidateEmailPerIP        ratelimit.BucketName = "OOBOTPValidateEmailPerIP"
	OOBOTPValidateSMSPerIP          ratelimit.BucketName = "OOBOTPValidateSMSPerIP"
	OOBOTPValidateEmailPerUserPerIP ratelimit.BucketName = "OOBOTPValidateEmailPerUserPerIP"
	OOBOTPValidateSMSPerUserPerIP   ratelimit.BucketName = "OOBOTPValidateSMSPerUserPerIP"
	AuthenticatePerIP               ratelimit.BucketName = "AuthenticatePerIP"
	AuthenticatePerUserPerIP        ratelimit.BucketName = "AuthenticatePerUserPerIP"
)

type kindOOBOTP struct {
	config  *config.AppConfig
	channel model.AuthenticatorOOBChannel
	purpose Purpose
}

func KindOOBOTP(config *config.AppConfig, channel model.AuthenticatorOOBChannel) Kind {
	return kindOOBOTP{config: config, channel: channel, purpose: PurposeOOBOTP}
}

var _ KindFactory = KindOOBOTP

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
		selectByChannel(k.channel,
			OOBOTPTriggerEmailPerIP,
			OOBOTPTriggerSMSPerIP,
		),
		string(k.purpose), ip)
}

func (k kindOOBOTP) RateLimitTriggerPerUser(userID string) ratelimit.BucketSpec {
	return ratelimit.NewBucketSpec(
		selectByChannel(k.channel,
			k.config.Authentication.RateLimits.OOBOTP.Email.TriggerPerUser,
			k.config.Authentication.RateLimits.OOBOTP.SMS.TriggerPerUser,
		),
		selectByChannel(k.channel,
			OOBOTPTriggerEmailPerUser,
			OOBOTPTriggerSMSPerUser,
		),
		string(k.purpose), userID)
}

func (k kindOOBOTP) RateLimitTriggerCooldown(target string) ratelimit.BucketSpec {
	return ratelimit.NewCooldownSpec(
		selectByChannel(k.channel,
			OOBOTPCooldownEmail,
			OOBOTPCooldownSMS,
		),
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
		return ratelimit.NewBucketSpec(
			k.config.Authentication.RateLimits.General.PerIP,
			AuthenticatePerIP,
			// purpose is omitted intentionally.
			ip,
		)
	}

	return ratelimit.NewBucketSpec(
		config,
		selectByChannel(k.channel,
			OOBOTPValidateEmailPerIP,
			OOBOTPValidateSMSPerIP,
		),
		string(k.purpose),
		ip,
	)
}

func (k kindOOBOTP) RateLimitValidatePerUserPerIP(userID string, ip string) ratelimit.BucketSpec {
	config := selectByChannel(k.channel,
		k.config.Authentication.RateLimits.OOBOTP.Email.ValidatePerUserPerIP,
		k.config.Authentication.RateLimits.OOBOTP.SMS.ValidatePerUserPerIP,
	)
	if config.Enabled == nil {
		return ratelimit.NewBucketSpec(
			k.config.Authentication.RateLimits.General.PerUserPerIP,
			AuthenticatePerUserPerIP,
			// purpose is omitted intentionally.
			userID,
			ip,
		)
	}

	return ratelimit.NewBucketSpec(
		config,
		selectByChannel(k.channel,
			OOBOTPValidateEmailPerUserPerIP,
			OOBOTPValidateSMSPerUserPerIP,
		),
		string(k.purpose),
		userID,
		ip,
	)
}

func (k kindOOBOTP) RevocationMaxFailedAttempts() int {
	return selectByChannel(k.channel,
		k.config.Authentication.RateLimits.OOBOTP.Email.MaxFailedAttemptsRevokeOTP,
		k.config.Authentication.RateLimits.OOBOTP.SMS.MaxFailedAttemptsRevokeOTP,
	)
}
