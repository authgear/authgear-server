package otp

import (
	"time"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/ratelimit"
	"github.com/authgear/authgear-server/pkg/util/secretcode"
)

type secretCode interface {
	Generate() string
	Compare(string, string) bool
}

type kindOOBOTP struct {
	config  *config.AppConfig
	channel model.AuthenticatorOOBChannel
	form    Form
}

func KindOOBOTP(config *config.AppConfig, channel model.AuthenticatorOOBChannel, form Form) Kind {
	return kindOOBOTP{config: config, channel: channel, form: form}
}

func (k kindOOBOTP) Purpose() string {
	return "oob-otp"
}

func (k kindOOBOTP) AllowLookupByCode() bool {
	return k.form == FormLink
}

func (k kindOOBOTP) code() secretCode {
	switch k.form {
	case FormCode:
		return secretcode.OOBOTPSecretCode
	case FormLink:
		return secretcode.LoginLinkOTPSecretCode
	default:
		panic("unexpected form: " + k.form)
	}
}

func (k kindOOBOTP) GenerateCode() string {
	return k.code().Generate()
}

func (k kindOOBOTP) VerifyCode(input string, expected string) bool {
	return k.code().Compare(input, expected)
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
