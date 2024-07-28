package otp

import (
	"time"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/ratelimit"
)

const PurposeOOBOTP Purpose = "oob-otp"

const (
	OOBOTPTriggerEmailPerIP            ratelimit.BucketName = "OOBOTPTriggerEmailPerIP"
	OOBOTPTriggerSMSPerIP              ratelimit.BucketName = "OOBOTPTriggerSMSPerIP"
	OOBOTPTriggerWhatsappPerIP         ratelimit.BucketName = "OOBOTPTriggerWhatsappPerIP"
	OOBOTPTriggerEmailPerUser          ratelimit.BucketName = "OOBOTPTriggerEmailPerUser"
	OOBOTPTriggerSMSPerUser            ratelimit.BucketName = "OOBOTPTriggerSMSPerUser"
	OOBOTPTriggerWhatsappPerUser       ratelimit.BucketName = "OOBOTPTriggerWhatsappPerUser"
	OOBOTPCooldownEmail                ratelimit.BucketName = "OOBOTPCooldownEmail"
	OOBOTPCooldownSMS                  ratelimit.BucketName = "OOBOTPCooldownSMS"
	OOBOTPCooldownWhatsapp             ratelimit.BucketName = "OOBOTPCooldownWhatsapp"
	OOBOTPValidateEmailPerIP           ratelimit.BucketName = "OOBOTPValidateEmailPerIP"
	OOBOTPValidateSMSPerIP             ratelimit.BucketName = "OOBOTPValidateSMSPerIP"
	OOBOTPValidateWhatsappPerIP        ratelimit.BucketName = "OOBOTPValidateWhatsappPerIP"
	OOBOTPValidateEmailPerUserPerIP    ratelimit.BucketName = "OOBOTPValidateEmailPerUserPerIP"
	OOBOTPValidateSMSPerUserPerIP      ratelimit.BucketName = "OOBOTPValidateSMSPerUserPerIP"
	OOBOTPValidateWhatsappPerUserPerIP ratelimit.BucketName = "OOBOTPValidateWhatsappPerUserPerIP"
	AuthenticatePerIP                  ratelimit.BucketName = "AuthenticatePerIP"
	AuthenticatePerUserPerIP           ratelimit.BucketName = "AuthenticatePerUserPerIP"
)

type kindOOBOTP struct {
	config  *config.AppConfig
	channel model.AuthenticatorOOBChannel
	purpose Purpose
	form 	  Form
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
	panic("unknown authenticator oob otp form");
}

func (k kindOOBOTP) RateLimitTriggerPerIP(ip string) ratelimit.BucketSpec {
	return ratelimit.NewBucketSpec(
		selectByChannel(k.channel,
			k.config.Authentication.RateLimits.OOBOTP.Email.TriggerPerIP,
			k.config.Authentication.RateLimits.OOBOTP.SMS.TriggerPerIP,
			k.config.Authentication.RateLimits.OOBOTP.SMS.TriggerPerIP,
		),
		selectByChannel(k.channel,
			OOBOTPTriggerEmailPerIP,
			OOBOTPTriggerSMSPerIP,
			OOBOTPTriggerWhatsappPerIP,
		),
		string(k.purpose), ip)
}

func (k kindOOBOTP) RateLimitTriggerPerUser(userID string) ratelimit.BucketSpec {
	return ratelimit.NewBucketSpec(
		selectByChannel(k.channel,
			k.config.Authentication.RateLimits.OOBOTP.Email.TriggerPerUser,
			k.config.Authentication.RateLimits.OOBOTP.SMS.TriggerPerUser,
			k.config.Authentication.RateLimits.OOBOTP.SMS.TriggerPerUser,
		),
		selectByChannel(k.channel,
			OOBOTPTriggerEmailPerUser,
			OOBOTPTriggerSMSPerUser,
			OOBOTPTriggerWhatsappPerUser,
		),
		string(k.purpose), userID)
}

func (k kindOOBOTP) RateLimitTriggerCooldown(target string) ratelimit.BucketSpec {
	return ratelimit.NewCooldownSpec(
		selectByChannel(k.channel,
			OOBOTPCooldownEmail,
			OOBOTPCooldownSMS,
			OOBOTPCooldownWhatsapp,
		),
		selectByChannel(k.channel,
			k.config.Authentication.RateLimits.OOBOTP.Email.TriggerCooldown,
			k.config.Authentication.RateLimits.OOBOTP.SMS.TriggerCooldown,
			k.config.Authentication.RateLimits.OOBOTP.SMS.TriggerCooldown,
		).Duration(),
		string(k.purpose), target,
	)
}

func (k kindOOBOTP) RateLimitValidatePerIP(ip string) ratelimit.BucketSpec {
	config := selectByChannel(k.channel,
		k.config.Authentication.RateLimits.OOBOTP.Email.ValidatePerIP,
		k.config.Authentication.RateLimits.OOBOTP.SMS.ValidatePerIP,
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
			OOBOTPValidateWhatsappPerIP,
		),
		string(k.purpose),
		ip,
	)
}

func (k kindOOBOTP) RateLimitValidatePerUserPerIP(userID string, ip string) ratelimit.BucketSpec {
	config := selectByChannel(k.channel,
		k.config.Authentication.RateLimits.OOBOTP.Email.ValidatePerUserPerIP,
		k.config.Authentication.RateLimits.OOBOTP.SMS.ValidatePerUserPerIP,
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
			OOBOTPValidateWhatsappPerUserPerIP,
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
		k.config.Authentication.RateLimits.OOBOTP.SMS.MaxFailedAttemptsRevokeOTP,
	)
}
