package ratelimit

import (
	"fmt"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/config"
)

type RateLimit string

const (
	// Authentication rate limits
	RateLimitAuthenticationGeneral             RateLimit = "authentication.general"
	RateLimitAuthenticationPassword            RateLimit = "authentication.password"
	RateLimitAuthenticationOOBOTPEmailTrigger  RateLimit = "authentication.oob_otp.email.trigger"
	RateLimitAuthenticationOOBOTPEmailValidate RateLimit = "authentication.oob_otp.email.validate"
	RateLimitAuthenticationOOBOTPSMSTrigger    RateLimit = "authentication.oob_otp.sms.trigger"
	RateLimitAuthenticationOOBOTPSMSValidate   RateLimit = "authentication.oob_otp.sms.validate"
	RateLimitAuthenticationTOTP                RateLimit = "authentication.totp"
	RateLimitAuthenticationRecoveryCode        RateLimit = "authentication.recovery_code"
	RateLimitAuthenticationDeviceToken         RateLimit = "authentication.device_token"
	RateLimitAuthenticationPasskey             RateLimit = "authentication.passkey"
	RateLimitAuthenticationSIWE                RateLimit = "authentication.siwe"
	RateLimitAuthenticationSignup              RateLimit = "authentication.signup"
	RateLimitAuthenticationSignupAnonymous     RateLimit = "authentication.signup_anonymous"
	RateLimitAuthenticationAccountEnumeration  RateLimit = "authentication.account_enumeration"

	// Features rate limits
	RateLimitVerificationEmailTrigger    RateLimit = "verification.email.trigger"
	RateLimitVerificationEmailValidate   RateLimit = "verification.email.validate"
	RateLimitVerificationSMSTrigger      RateLimit = "verification.sms.trigger"
	RateLimitVerificationSMSValidate     RateLimit = "verification.sms.validate"
	RateLimitForgotPasswordEmailTrigger  RateLimit = "forgot_password.email.trigger"
	RateLimitForgotPasswordEmailValidate RateLimit = "forgot_password.email.validate"
	RateLimitForgotPasswordSMSTrigger    RateLimit = "forgot_password.sms.trigger"
	RateLimitForgotPasswordSMSValidate   RateLimit = "forgot_password.sms.validate"

	// Messaging rate limits
	RateLimitMessagingSMS   RateLimit = "messaging.sms"
	RateLimitMessagingEmail RateLimit = "messaging.email"
)

const (
	SignupAnonymousPerIP    BucketName = "SignupAnonymousPerIP"
	SignupPerIP             BucketName = "SignupPerIP"
	AccountEnumerationPerIP BucketName = "AccountEnumerationPerIP"

	OOBOTPTriggerEmailPerIP            BucketName = "OOBOTPTriggerEmailPerIP"
	OOBOTPTriggerSMSPerIP              BucketName = "OOBOTPTriggerSMSPerIP"
	OOBOTPTriggerWhatsappPerIP         BucketName = "OOBOTPTriggerWhatsappPerIP"
	OOBOTPTriggerEmailPerUser          BucketName = "OOBOTPTriggerEmailPerUser"
	OOBOTPTriggerSMSPerUser            BucketName = "OOBOTPTriggerSMSPerUser"
	OOBOTPTriggerWhatsappPerUser       BucketName = "OOBOTPTriggerWhatsappPerUser"
	OOBOTPCooldownEmail                BucketName = "OOBOTPCooldownEmail"
	OOBOTPCooldownSMS                  BucketName = "OOBOTPCooldownSMS"
	OOBOTPCooldownWhatsapp             BucketName = "OOBOTPCooldownWhatsapp"
	OOBOTPValidateEmailPerIP           BucketName = "OOBOTPValidateEmailPerIP"
	OOBOTPValidateSMSPerIP             BucketName = "OOBOTPValidateSMSPerIP"
	OOBOTPValidateWhatsappPerIP        BucketName = "OOBOTPValidateWhatsappPerIP"
	OOBOTPValidateEmailPerUserPerIP    BucketName = "OOBOTPValidateEmailPerUserPerIP"
	OOBOTPValidateSMSPerUserPerIP      BucketName = "OOBOTPValidateSMSPerUserPerIP"
	OOBOTPValidateWhatsappPerUserPerIP BucketName = "OOBOTPValidateWhatsappPerUserPerIP"

	VerificationTriggerEmailPerIP      BucketName = "VerificationTriggerEmailPerIP"
	VerificationTriggerSMSPerIP        BucketName = "VerificationTriggerSMSPerIP"
	VerificationTriggerWhatsappPerIP   BucketName = "VerificationTriggerWhatsappPerIP"
	VerificationTriggerEmailPerUser    BucketName = "VerificationTriggerEmailPerUser"
	VerificationTriggerSMSPerUser      BucketName = "VerificationTriggerSMSPerUser"
	VerificationTriggerWhatsappPerUser BucketName = "VerificationTriggerWhatsappPerUser"
	VerificationCooldownEmail          BucketName = "VerificationCooldownEmail"
	VerificationCooldownSMS            BucketName = "VerificationCooldownSMS"
	VerificationCooldownWhatsapp       BucketName = "VerificationCooldownWhatsapp"
	VerificationValidateEmailPerIP     BucketName = "VerificationValidateEmailPerIP"
	VerificationValidateSMSPerIP       BucketName = "VerificationValidateSMSPerIP"
	VerificationValidateWhatsappPerIP  BucketName = "VerificationValidateWhatsappPerIP"

	VerifyPasswordPerIP              BucketName = "VerifyPasswordPerIP"
	VerifyPasswordPerUserPerIP       BucketName = "VerifyPasswordPerUserPerIP"
	VerifyTOTPPerIP                  BucketName = "VerifyTOTPPerIP"
	VerifyTOTPPerUserPerIP           BucketName = "VerifyTOTPPerUserPerIP"
	VerifyPasskeyPerIP               BucketName = "VerifyPasskeyPerIP"
	VerifyRecoveryCodePerIP          BucketName = "VerifyRecoveryCodePerIP"
	VerifyRecoveryCodePerUserPerIP   BucketName = "VerifyRecoveryCodePerUserPerIP"
	VerifyDeviceTokenPerIP           BucketName = "VerifyDeviceTokenPerIP"
	VerifyDeviceTokenPerUserPerIP    BucketName = "VerifyDeviceTokenPerUserPerIP"
	VerifySIWEPerIP                  BucketName = "VerifySIWEPerIP"
	ForgotPasswordTriggerEmailPerIP  BucketName = "ForgotPasswordTriggerEmailPerIP"
	ForgotPasswordValidateEmailPerIP BucketName = "ForgotPasswordValidateEmailPerIP"
	ForgotPasswordTriggerSMSPerIP    BucketName = "ForgotPasswordTriggerSMSPerIP"
	ForgotPasswordValidateSMSPerIP   BucketName = "ForgotPasswordValidateSMSPerIP"

	MessagingSMSPerIP   BucketName = "MessagingSMSPerIP"
	MessagingEmailPerIP BucketName = "MessagingEmailPerIP"

	PresignImageUploadRequestPerUser BucketName = "PresignImageUploadRequestPerUser"
)

func (n RateLimit) resolvePerIP(cfg *config.AppConfig) *config.RateLimitConfig {
	get := func(c *config.RateLimitConfig, fallback *config.RateLimitConfig) *config.RateLimitConfig {
		if c != nil && c.IsEnabled() {
			return c
		}
		if fallback != nil && fallback.IsEnabled() {
			return fallback
		}
		return nil
	}
	switch n {
	// Authentication
	case RateLimitAuthenticationGeneral:
		return get(cfg.Authentication.RateLimits.General.PerIP, nil)
	case RateLimitAuthenticationPassword:
		return get(cfg.Authentication.RateLimits.Password.PerIP, cfg.Authentication.RateLimits.General.PerIP)
	case RateLimitAuthenticationOOBOTPEmailTrigger:
		return get(cfg.Authentication.RateLimits.OOBOTP.Email.TriggerPerIP, nil)
	case RateLimitAuthenticationOOBOTPEmailValidate:
		return get(cfg.Authentication.RateLimits.OOBOTP.Email.ValidatePerIP, cfg.Authentication.RateLimits.General.PerIP)
	case RateLimitAuthenticationOOBOTPSMSTrigger:
		return get(cfg.Authentication.RateLimits.OOBOTP.SMS.TriggerPerIP, nil)
	case RateLimitAuthenticationOOBOTPSMSValidate:
		return get(cfg.Authentication.RateLimits.OOBOTP.SMS.ValidatePerIP, cfg.Authentication.RateLimits.General.PerIP)
	case RateLimitAuthenticationTOTP:
		return get(cfg.Authentication.RateLimits.TOTP.PerIP, cfg.Authentication.RateLimits.General.PerIP)
	case RateLimitAuthenticationRecoveryCode:
		return get(cfg.Authentication.RateLimits.RecoveryCode.PerIP, cfg.Authentication.RateLimits.General.PerIP)
	case RateLimitAuthenticationDeviceToken:
		return get(cfg.Authentication.RateLimits.DeviceToken.PerIP, cfg.Authentication.RateLimits.General.PerIP)
	case RateLimitAuthenticationPasskey:
		return get(cfg.Authentication.RateLimits.Passkey.PerIP, cfg.Authentication.RateLimits.General.PerIP)
	case RateLimitAuthenticationSIWE:
		return get(cfg.Authentication.RateLimits.SIWE.PerIP, cfg.Authentication.RateLimits.General.PerIP)
	case RateLimitAuthenticationSignup:
		return get(cfg.Authentication.RateLimits.Signup.PerIP, nil)
	case RateLimitAuthenticationSignupAnonymous:
		return get(cfg.Authentication.RateLimits.SignupAnonymous.PerIP, nil)
	case RateLimitAuthenticationAccountEnumeration:
		return get(cfg.Authentication.RateLimits.AccountEnumeration.PerIP, nil)

	// Features
	case RateLimitVerificationEmailTrigger:
		return get(cfg.Verification.RateLimits.Email.TriggerPerIP, nil)
	case RateLimitVerificationEmailValidate:
		return get(cfg.Verification.RateLimits.Email.ValidatePerIP, nil)
	case RateLimitVerificationSMSTrigger:
		return get(cfg.Verification.RateLimits.SMS.TriggerPerIP, nil)
	case RateLimitVerificationSMSValidate:
		return get(cfg.Verification.RateLimits.SMS.ValidatePerIP, nil)
	case RateLimitForgotPasswordEmailTrigger:
		return get(cfg.ForgotPassword.RateLimits.Email.TriggerPerIP, nil)
	case RateLimitForgotPasswordEmailValidate:
		return get(cfg.ForgotPassword.RateLimits.Email.ValidatePerIP, nil)
	case RateLimitForgotPasswordSMSTrigger:
		return get(cfg.ForgotPassword.RateLimits.SMS.TriggerPerIP, nil)
	case RateLimitForgotPasswordSMSValidate:
		return get(cfg.ForgotPassword.RateLimits.SMS.ValidatePerIP, nil)

	// Messaging
	case RateLimitMessagingSMS:
		return get(cfg.Messaging.RateLimits.SMSPerIP, nil)
	case RateLimitMessagingEmail:
		return get(cfg.Messaging.RateLimits.EmailPerIP, nil)
	}
	return nil
}

func (n RateLimit) resolvePerUser(cfg *config.AppConfig) *config.RateLimitConfig {
	get := func(c *config.RateLimitConfig, fallback *config.RateLimitConfig) *config.RateLimitConfig {
		if c != nil && c.IsEnabled() {
			return c
		}
		if fallback != nil && fallback.IsEnabled() {
			return fallback
		}
		return nil
	}
	switch n {
	case RateLimitAuthenticationOOBOTPEmailTrigger:
		return get(cfg.Authentication.RateLimits.OOBOTP.Email.TriggerPerUser, nil)
	case RateLimitAuthenticationOOBOTPSMSTrigger:
		return get(cfg.Authentication.RateLimits.OOBOTP.SMS.TriggerPerUser, nil)
	}
	return nil
}

func (n RateLimit) resolvePerUserPerIP(cfg *config.AppConfig) *config.RateLimitConfig {
	get := func(c *config.RateLimitConfig, fallback *config.RateLimitConfig) *config.RateLimitConfig {
		if c != nil && c.IsEnabled() {
			return c
		}
		if fallback != nil && fallback.IsEnabled() {
			return fallback
		}
		return nil
	}
	switch n {
	case RateLimitAuthenticationGeneral:
		return get(cfg.Authentication.RateLimits.General.PerUserPerIP, nil)
	case RateLimitAuthenticationPassword:
		return get(cfg.Authentication.RateLimits.Password.PerUserPerIP, cfg.Authentication.RateLimits.General.PerUserPerIP)
	case RateLimitAuthenticationOOBOTPEmailValidate:
		return get(cfg.Authentication.RateLimits.OOBOTP.Email.ValidatePerUserPerIP, cfg.Authentication.RateLimits.General.PerUserPerIP)
	case RateLimitAuthenticationOOBOTPSMSValidate:
		return get(cfg.Authentication.RateLimits.OOBOTP.SMS.ValidatePerUserPerIP, cfg.Authentication.RateLimits.General.PerUserPerIP)
	case RateLimitAuthenticationTOTP:
		return get(cfg.Authentication.RateLimits.TOTP.PerUserPerIP, cfg.Authentication.RateLimits.General.PerUserPerIP)
	case RateLimitAuthenticationRecoveryCode:
		return get(cfg.Authentication.RateLimits.RecoveryCode.PerUserPerIP, cfg.Authentication.RateLimits.General.PerUserPerIP)
	case RateLimitAuthenticationDeviceToken:
		return get(cfg.Authentication.RateLimits.DeviceToken.PerUserPerIP, cfg.Authentication.RateLimits.General.PerUserPerIP)
	}
	return nil
}

type ResolveBucketSpecOptions struct {
	IPAddress string
	UserID    string
	Purpose   string
	Channel   model.AuthenticatorOOBChannel
}

func selectByChannel[T any](channel model.AuthenticatorOOBChannel, email T, sms T, whatsapp T) T {
	switch channel {
	case model.AuthenticatorOOBChannelEmail:
		return email
	case model.AuthenticatorOOBChannelSMS:
		return sms
	case model.AuthenticatorOOBChannelWhatsapp:
		return whatsapp
	}
	panic(fmt.Errorf("invalid channel"))
}

func (r RateLimit) ResolveBucketSpecs(cfg *config.AppConfig, opts *ResolveBucketSpecOptions) []*BucketSpec {
	var specs []*BucketSpec

	resolvePerIPBucket := func(bucketName BucketName, args ...string) *BucketSpec {
		confPerIP := r.resolvePerIP(cfg)
		if confPerIP == nil {
			return &BucketSpecDisabled
		}
		spec := NewBucketSpec(confPerIP, bucketName, args...)
		return &spec
	}

	resolvePerUserBucket := func(bucketName BucketName, args ...string) *BucketSpec {
		confPerUser := r.resolvePerUser(cfg)
		if confPerUser == nil {
			return &BucketSpecDisabled
		}
		spec := NewBucketSpec(confPerUser, bucketName, args...)
		return &spec
	}

	resolvePerUserPerIPBucket := func(bucketName BucketName, args ...string) *BucketSpec {
		confPerUserPerIP := r.resolvePerUserPerIP(cfg)
		if confPerUserPerIP == nil {
			return &BucketSpecDisabled
		}
		spec := NewBucketSpec(confPerUserPerIP, bucketName, args...)
		return &spec
	}

	switch r {
	case RateLimitAuthenticationGeneral:
		// This rate limit is for fallback, so it should not be used directly
		panic(fmt.Errorf("%s has no buckets", r))
	case RateLimitAuthenticationPassword:
		specs = append(specs, resolvePerIPBucket(VerifyPasswordPerIP, opts.IPAddress))
		specs = append(specs, resolvePerUserPerIPBucket(VerifyPasswordPerUserPerIP, opts.UserID, opts.IPAddress))

	case RateLimitAuthenticationOOBOTPEmailTrigger:
		bucketNamePerIP := selectByChannel(opts.Channel,
			OOBOTPTriggerEmailPerIP,
			OOBOTPTriggerSMSPerIP,
			OOBOTPTriggerWhatsappPerIP,
		)
		specs = append(specs, resolvePerIPBucket(bucketNamePerIP, opts.Purpose, opts.IPAddress))

		bucketNamePerUser := selectByChannel(opts.Channel,
			OOBOTPTriggerEmailPerUser,
			OOBOTPTriggerSMSPerUser,
			OOBOTPTriggerWhatsappPerUser,
		)
		specs = append(specs, resolvePerUserBucket(bucketNamePerUser, opts.Purpose, opts.UserID))

	case RateLimitAuthenticationOOBOTPEmailValidate:
		bucketNamePerIP := selectByChannel(opts.Channel,
			OOBOTPValidateEmailPerIP,
			OOBOTPValidateSMSPerIP,
			OOBOTPValidateWhatsappPerIP,
		)
		specs = append(specs, resolvePerIPBucket(bucketNamePerIP, opts.Purpose, opts.IPAddress))

		bucketNamePerUserPerIP := selectByChannel(opts.Channel,
			OOBOTPValidateEmailPerUserPerIP,
			OOBOTPValidateSMSPerUserPerIP,
			OOBOTPValidateWhatsappPerUserPerIP,
		)
		specs = append(specs, resolvePerUserPerIPBucket(bucketNamePerUserPerIP, opts.Purpose, opts.UserID, opts.IPAddress))

	case RateLimitAuthenticationOOBOTPSMSTrigger:
		bucketNamePerIP := selectByChannel(opts.Channel,
			OOBOTPTriggerEmailPerIP,
			OOBOTPTriggerSMSPerIP,
			OOBOTPTriggerWhatsappPerIP,
		)
		specs = append(specs, resolvePerIPBucket(bucketNamePerIP, opts.Purpose, opts.IPAddress))

		bucketNamePerUser := selectByChannel(opts.Channel,
			OOBOTPTriggerEmailPerUser,
			OOBOTPTriggerSMSPerUser,
			OOBOTPTriggerWhatsappPerUser,
		)
		specs = append(specs, resolvePerUserBucket(bucketNamePerUser, opts.Purpose, opts.UserID))

	case RateLimitAuthenticationOOBOTPSMSValidate:
		bucketNamePerIP := selectByChannel(opts.Channel,
			OOBOTPValidateEmailPerIP,
			OOBOTPValidateSMSPerIP,
			OOBOTPValidateWhatsappPerIP,
		)
		specs = append(specs, resolvePerIPBucket(bucketNamePerIP, opts.Purpose, opts.IPAddress))

		bucketNamePerUserPerIP := selectByChannel(opts.Channel,
			OOBOTPValidateEmailPerUserPerIP,
			OOBOTPValidateSMSPerUserPerIP,
			OOBOTPValidateWhatsappPerUserPerIP,
		)
		specs = append(specs, resolvePerUserPerIPBucket(bucketNamePerUserPerIP, opts.Purpose, opts.UserID, opts.IPAddress))

	case RateLimitAuthenticationTOTP:
		specs = append(specs, resolvePerIPBucket(VerifyTOTPPerIP, opts.IPAddress))
		specs = append(specs, resolvePerUserPerIPBucket(VerifyTOTPPerUserPerIP, opts.UserID, opts.IPAddress))

	case RateLimitAuthenticationRecoveryCode:
		specs = append(specs, resolvePerIPBucket(VerifyRecoveryCodePerIP, opts.IPAddress))
		specs = append(specs, resolvePerUserPerIPBucket(VerifyRecoveryCodePerUserPerIP, opts.UserID, opts.IPAddress))

	case RateLimitAuthenticationDeviceToken:
		specs = append(specs, resolvePerIPBucket(VerifyDeviceTokenPerIP, opts.IPAddress))
		specs = append(specs, resolvePerUserPerIPBucket(VerifyDeviceTokenPerUserPerIP, opts.UserID, opts.IPAddress))

	case RateLimitAuthenticationPasskey:
		specs = append(specs, resolvePerIPBucket(VerifyPasskeyPerIP, opts.IPAddress))

	case RateLimitAuthenticationSIWE:
		specs = append(specs, resolvePerIPBucket(VerifySIWEPerIP, opts.IPAddress))

	case RateLimitAuthenticationSignup:
		specs = append(specs, resolvePerIPBucket(SignupPerIP, opts.IPAddress))

	case RateLimitAuthenticationSignupAnonymous:
		specs = append(specs, resolvePerIPBucket(SignupAnonymousPerIP, opts.IPAddress))

	case RateLimitAuthenticationAccountEnumeration:
		specs = append(specs, resolvePerIPBucket(AccountEnumerationPerIP, opts.IPAddress))

	case RateLimitVerificationEmailTrigger:
		bucketNamePerIP := selectByChannel(opts.Channel,
			VerificationTriggerEmailPerIP,
			VerificationTriggerSMSPerIP,
			VerificationTriggerWhatsappPerIP,
		)
		specs = append(specs, resolvePerIPBucket(bucketNamePerIP, opts.IPAddress))

		bucketNamePerUser := selectByChannel(opts.Channel,
			VerificationTriggerEmailPerUser,
			VerificationTriggerSMSPerUser,
			VerificationTriggerWhatsappPerUser,
		)
		specs = append(specs, resolvePerUserBucket(bucketNamePerUser, opts.UserID))

	case RateLimitVerificationEmailValidate:
		bucketNamePerIP := selectByChannel(opts.Channel,
			VerificationValidateEmailPerIP,
			VerificationValidateSMSPerIP,
			VerificationValidateWhatsappPerIP,
		)
		specs = append(specs, resolvePerIPBucket(bucketNamePerIP, opts.IPAddress))

	case RateLimitVerificationSMSTrigger:
		bucketNamePerIP := selectByChannel(opts.Channel,
			VerificationTriggerEmailPerIP,
			VerificationTriggerSMSPerIP,
			VerificationTriggerWhatsappPerIP,
		)
		specs = append(specs, resolvePerIPBucket(bucketNamePerIP, opts.IPAddress))

		bucketNamePerUser := selectByChannel(opts.Channel,
			VerificationTriggerEmailPerUser,
			VerificationTriggerSMSPerUser,
			VerificationTriggerWhatsappPerUser,
		)
		specs = append(specs, resolvePerUserBucket(bucketNamePerUser, opts.UserID))

	case RateLimitVerificationSMSValidate:
		bucketNamePerIP := selectByChannel(opts.Channel,
			VerificationValidateEmailPerIP,
			VerificationValidateSMSPerIP,
			VerificationValidateWhatsappPerIP,
		)
		specs = append(specs, resolvePerIPBucket(bucketNamePerIP, opts.IPAddress))

	case RateLimitForgotPasswordEmailTrigger:
		specs = append(specs, resolvePerIPBucket(ForgotPasswordTriggerEmailPerIP, opts.IPAddress))

	case RateLimitForgotPasswordEmailValidate:
		specs = append(specs, resolvePerIPBucket(ForgotPasswordValidateEmailPerIP, opts.IPAddress))

	case RateLimitForgotPasswordSMSTrigger:
		specs = append(specs, resolvePerIPBucket(ForgotPasswordTriggerSMSPerIP, opts.IPAddress))

	case RateLimitForgotPasswordSMSValidate:
		specs = append(specs, resolvePerIPBucket(ForgotPasswordValidateSMSPerIP, opts.IPAddress))

	case RateLimitMessagingSMS:
		specs = append(specs, resolvePerIPBucket(MessagingSMSPerIP, opts.IPAddress))

	case RateLimitMessagingEmail:
		specs = append(specs, resolvePerIPBucket(MessagingEmailPerIP, opts.IPAddress))
	}

	return specs
}
