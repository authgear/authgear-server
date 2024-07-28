package otp

import (
	"time"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/ratelimit"
)

const PurposeVerification Purpose = "verification"

const (
	VerificationTriggerEmailPerIP      ratelimit.BucketName = "VerificationTriggerEmailPerIP"
	VerificationTriggerSMSPerIP        ratelimit.BucketName = "VerificationTriggerSMSPerIP"
	VerificationTriggerWhatsappPerIP   ratelimit.BucketName = "VerificationTriggerWhatsappPerIP"
	VerificationTriggerEmailPerUser    ratelimit.BucketName = "VerificationTriggerEmailPerUser"
	VerificationTriggerSMSPerUser      ratelimit.BucketName = "VerificationTriggerSMSPerUser"
	VerificationTriggerWhatsappPerUser ratelimit.BucketName = "VerificationTriggerWhatsappPerUser"
	VerificationCooldownEmail          ratelimit.BucketName = "VerificationCooldownEmail"
	VerificationCooldownSMS            ratelimit.BucketName = "VerificationCooldownSMS"
	VerificationCooldownWhatsapp       ratelimit.BucketName = "VerificationCooldownWhatsapp"
	VerificationValidateEmailPerIP     ratelimit.BucketName = "VerificationValidateEmailPerIP"
	VerificationValidateSMSPerIP       ratelimit.BucketName = "VerificationValidateSMSPerIP"
	VerificationValidateWhatsappPerIP  ratelimit.BucketName = "VerificationValidateWhatsappPerIP"
)

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

func (k kindVerification) RateLimitTriggerPerIP(ip string) ratelimit.BucketSpec {
	return ratelimit.NewBucketSpec(
		selectByChannel(k.channel,
			k.config.Verification.RateLimits.Email.TriggerPerIP,
			k.config.Verification.RateLimits.SMS.TriggerPerIP,
			k.config.Verification.RateLimits.SMS.TriggerPerIP,
		),
		selectByChannel(k.channel,
			VerificationTriggerEmailPerIP,
			VerificationTriggerSMSPerIP,
			VerificationTriggerWhatsappPerIP,
		), ip)
}

func (k kindVerification) RateLimitTriggerPerUser(userID string) ratelimit.BucketSpec {
	return ratelimit.NewBucketSpec(
		selectByChannel(k.channel,
			k.config.Verification.RateLimits.Email.TriggerPerUser,
			k.config.Verification.RateLimits.SMS.TriggerPerUser,
			k.config.Verification.RateLimits.SMS.TriggerPerUser,
		),
		selectByChannel(k.channel,
			VerificationTriggerEmailPerUser,
			VerificationTriggerSMSPerUser,
			VerificationTriggerWhatsappPerUser,
		), userID)
}

func (k kindVerification) RateLimitTriggerCooldown(target string) ratelimit.BucketSpec {
	return ratelimit.NewCooldownSpec(
		selectByChannel(k.channel,
			VerificationCooldownEmail,
			VerificationCooldownSMS,
			VerificationCooldownWhatsapp,
		),
		selectByChannel(k.channel,
			k.config.Verification.RateLimits.Email.TriggerCooldown,
			k.config.Verification.RateLimits.SMS.TriggerCooldown,
			k.config.Verification.RateLimits.SMS.TriggerCooldown,
		).Duration(),
		target,
	)
}

func (k kindVerification) RateLimitValidatePerIP(ip string) ratelimit.BucketSpec {
	return ratelimit.NewBucketSpec(
		selectByChannel(k.channel,
			k.config.Verification.RateLimits.Email.ValidatePerIP,
			k.config.Verification.RateLimits.SMS.ValidatePerIP,
			k.config.Verification.RateLimits.SMS.ValidatePerIP,
		),
		selectByChannel(k.channel,
			VerificationValidateEmailPerIP,
			VerificationValidateSMSPerIP,
			VerificationValidateWhatsappPerIP,
		), ip)
}

func (k kindVerification) RateLimitValidatePerUserPerIP(userID string, ip string) ratelimit.BucketSpec {
	return ratelimit.BucketSpecDisabled
}

func (k kindVerification) RevocationMaxFailedAttempts() int {
	return selectByChannel(k.channel,
		k.config.Verification.RateLimits.Email.MaxFailedAttemptsRevokeOTP,
		k.config.Verification.RateLimits.SMS.MaxFailedAttemptsRevokeOTP,
		k.config.Verification.RateLimits.SMS.MaxFailedAttemptsRevokeOTP,
	)
}
