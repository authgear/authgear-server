package otp

import (
	"time"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/ratelimit"
)

const PurposeForgotPassword Purpose = "forgot-password"

const (
	ForgotPasswordTriggerEmailPerIP     ratelimit.BucketName = "ForgotPasswordTriggerEmailPerIP"
	ForgotPasswordTriggerSMSPerIP       ratelimit.BucketName = "ForgotPasswordTriggerSMSPerIP"
	ForgotPasswordTriggerWhatsappPerIP  ratelimit.BucketName = "ForgotPasswordTriggerWhatsappPerIP"
	ForgotPasswordCooldownEmail         ratelimit.BucketName = "ForgotPasswordCooldownEmail"
	ForgotPasswordCooldownSMS           ratelimit.BucketName = "ForgotPasswordCooldownSMS"
	ForgotPasswordCooldownWhatsapp      ratelimit.BucketName = "ForgotPasswordCooldownWhatsapp"
	ForgotPasswordValidateEmailPerIP    ratelimit.BucketName = "ForgotPasswordValidateEmailPerIP"
	ForgotPasswordValidateSMSPerIP      ratelimit.BucketName = "ForgotPasswordValidateSMSPerIP"
	ForgotPasswordValidateWhatsappPerIP ratelimit.BucketName = "ForgotPasswordValidateWhatsappPerIP"
)

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

func (k kindForgotPassword) RateLimitTriggerPerIP(ip string) ratelimit.BucketSpec {
	return ratelimit.NewBucketSpec(
		selectByChannel(
			k.channel,
			k.config.ForgotPassword.RateLimits.Email.TriggerPerIP,
			k.config.ForgotPassword.RateLimits.SMS.TriggerPerIP,
			k.config.ForgotPassword.RateLimits.SMS.TriggerPerIP,
		),
		selectByChannel(
			k.channel,
			ForgotPasswordTriggerEmailPerIP,
			ForgotPasswordTriggerSMSPerIP,
			ForgotPasswordTriggerWhatsappPerIP,
		),
		ip,
	)
}

func (k kindForgotPassword) RateLimitTriggerPerUser(userID string) ratelimit.BucketSpec {
	return ratelimit.BucketSpecDisabled
}

func (k kindForgotPassword) RateLimitTriggerCooldown(target string) ratelimit.BucketSpec {
	return ratelimit.NewCooldownSpec(
		selectByChannel(
			k.channel,
			ForgotPasswordCooldownEmail,
			ForgotPasswordCooldownSMS,
			ForgotPasswordCooldownWhatsapp,
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

func (k kindForgotPassword) RateLimitValidatePerIP(ip string) ratelimit.BucketSpec {
	return ratelimit.NewBucketSpec(
		selectByChannel(
			k.channel,
			k.config.ForgotPassword.RateLimits.Email.ValidatePerIP,
			k.config.ForgotPassword.RateLimits.SMS.ValidatePerIP,
			k.config.ForgotPassword.RateLimits.SMS.ValidatePerIP,
		),
		selectByChannel(
			k.channel,
			ForgotPasswordValidateEmailPerIP,
			ForgotPasswordValidateSMSPerIP,
			ForgotPasswordValidateWhatsappPerIP,
		),
		ip,
	)
}

func (k kindForgotPassword) RateLimitValidatePerUserPerIP(userID string, ip string) ratelimit.BucketSpec {
	return ratelimit.BucketSpecDisabled
}

func (k kindForgotPassword) RevocationMaxFailedAttempts() int {
	return 0
}
