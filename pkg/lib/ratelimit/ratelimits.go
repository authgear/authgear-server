package ratelimit

import (
	"context"
	"fmt"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/config"
)

type RateLimitGroup string

const (
	// Authentication rate limits
	RateLimitAuthenticationGeneral             RateLimitGroup = "authentication.general"
	RateLimitAuthenticationPassword            RateLimitGroup = "authentication.password"
	RateLimitAuthenticationOOBOTPEmailTrigger  RateLimitGroup = "authentication.oob_otp.email.trigger"
	RateLimitAuthenticationOOBOTPEmailValidate RateLimitGroup = "authentication.oob_otp.email.validate"
	RateLimitAuthenticationOOBOTPSMSTrigger    RateLimitGroup = "authentication.oob_otp.sms.trigger"
	RateLimitAuthenticationOOBOTPSMSValidate   RateLimitGroup = "authentication.oob_otp.sms.validate"
	RateLimitAuthenticationTOTP                RateLimitGroup = "authentication.totp"
	RateLimitAuthenticationRecoveryCode        RateLimitGroup = "authentication.recovery_code"
	RateLimitAuthenticationDeviceToken         RateLimitGroup = "authentication.device_token"
	RateLimitAuthenticationPasskey             RateLimitGroup = "authentication.passkey"
	RateLimitAuthenticationSIWE                RateLimitGroup = "authentication.siwe"
	RateLimitAuthenticationSignup              RateLimitGroup = "authentication.signup"
	RateLimitAuthenticationSignupAnonymous     RateLimitGroup = "authentication.signup_anonymous"
	RateLimitAuthenticationAccountEnumeration  RateLimitGroup = "authentication.account_enumeration"

	// Features rate limits
	RateLimitVerificationEmailTrigger    RateLimitGroup = "verification.email.trigger"
	RateLimitVerificationEmailValidate   RateLimitGroup = "verification.email.validate"
	RateLimitVerificationSMSTrigger      RateLimitGroup = "verification.sms.trigger"
	RateLimitVerificationSMSValidate     RateLimitGroup = "verification.sms.validate"
	RateLimitForgotPasswordEmailTrigger  RateLimitGroup = "forgot_password.email.trigger"
	RateLimitForgotPasswordEmailValidate RateLimitGroup = "forgot_password.email.validate"
	RateLimitForgotPasswordSMSTrigger    RateLimitGroup = "forgot_password.sms.trigger"
	RateLimitForgotPasswordSMSValidate   RateLimitGroup = "forgot_password.sms.validate"

	// Messaging rate limits
	RateLimitMessagingSMS   RateLimitGroup = "messaging.sms"
	RateLimitMessagingEmail RateLimitGroup = "messaging.email"

	// Token endpoint rate limits
	RateLimitOAuthTokenGeneral           RateLimitGroup = "oauth.token.general"            // #nosec G101
	RateLimitOAuthTokenClientCredentials RateLimitGroup = "oauth.token.client_credentials" // #nosec G101
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

	VerifyPasswordPerIP            BucketName = "VerifyPasswordPerIP"
	VerifyPasswordPerUserPerIP     BucketName = "VerifyPasswordPerUserPerIP"
	VerifyTOTPPerIP                BucketName = "VerifyTOTPPerIP"
	VerifyTOTPPerUserPerIP         BucketName = "VerifyTOTPPerUserPerIP"
	VerifyPasskeyPerIP             BucketName = "VerifyPasskeyPerIP"
	VerifyRecoveryCodePerIP        BucketName = "VerifyRecoveryCodePerIP"
	VerifyRecoveryCodePerUserPerIP BucketName = "VerifyRecoveryCodePerUserPerIP"
	VerifyDeviceTokenPerIP         BucketName = "VerifyDeviceTokenPerIP"
	VerifyDeviceTokenPerUserPerIP  BucketName = "VerifyDeviceTokenPerUserPerIP"
	VerifySIWEPerIP                BucketName = "VerifySIWEPerIP"

	ForgotPasswordTriggerEmailPerIP    BucketName = "ForgotPasswordTriggerEmailPerIP"
	ForgotPasswordTriggerSMSPerIP      BucketName = "ForgotPasswordTriggerSMSPerIP"
	ForgotPasswordTriggerWhatsappPerIP BucketName = "ForgotPasswordTriggerWhatsappPerIP"

	ForgotPasswordValidateEmailPerIP    BucketName = "ForgotPasswordValidateEmailPerIP"
	ForgotPasswordValidateSMSPerIP      BucketName = "ForgotPasswordValidateSMSPerIP"
	ForgotPasswordValidateWhatsappPerIP BucketName = "ForgotPasswordValidateWhatsappPerIP"

	ForgotPasswordCooldownEmail    BucketName = "ForgotPasswordCooldownEmail"
	ForgotPasswordCooldownSMS      BucketName = "ForgotPasswordCooldownSMS"
	ForgotPasswordCooldownWhatsapp BucketName = "ForgotPasswordCooldownWhatsapp"

	MessagingSMS          BucketName = "MessagingSMS"
	MessagingSMSPerIP     BucketName = "MessagingSMSPerIP"
	MessagingSMSPerTarget BucketName = "MessagingSMSPerTarget"

	MessagingEmail          BucketName = "MessagingEmail"
	MessagingEmailPerIP     BucketName = "MessagingEmailPerIP"
	MessagingEmailPerTarget BucketName = "MessagingEmailPerTarget"

	PresignImageUploadRequestPerUser BucketName = "PresignImageUploadRequestPerUser"

	OAuthTokenPerIP                       BucketName = "OAuthTokenPerIP"   // #nosec G101
	OAuthTokenPerUser                     BucketName = "OAuthTokenPerUser" // #nosec G101
	OAuthTokenClientCredentialsPerClient  BucketName = "OAuthTokenClientCredentialsPerClient"
	OAuthTokenClientCredentialsPerProject BucketName = "OAuthTokenClientCredentialsPerProject"
)

func (n RateLimitGroup) resolvePerIP(cfg *config.AppConfig, featureCfg *config.FeatureConfig) *config.RateLimitConfig {
	switch n {
	// Authentication
	case RateLimitAuthenticationGeneral:
		return resolveConfig(cfg.Authentication.RateLimits.General.PerIP, nil, nil)
	case RateLimitAuthenticationPassword:
		return resolveConfig(cfg.Authentication.RateLimits.Password.PerIP, cfg.Authentication.RateLimits.General.PerIP, nil)
	case RateLimitAuthenticationOOBOTPEmailTrigger:
		return resolveConfig(cfg.Authentication.RateLimits.OOBOTP.Email.TriggerPerIP, nil, nil)
	case RateLimitAuthenticationOOBOTPEmailValidate:
		return resolveConfig(cfg.Authentication.RateLimits.OOBOTP.Email.ValidatePerIP, cfg.Authentication.RateLimits.General.PerIP, nil)
	case RateLimitAuthenticationOOBOTPSMSTrigger:
		return resolveConfig(cfg.Authentication.RateLimits.OOBOTP.SMS.TriggerPerIP, nil, nil)
	case RateLimitAuthenticationOOBOTPSMSValidate:
		return resolveConfig(cfg.Authentication.RateLimits.OOBOTP.SMS.ValidatePerIP, cfg.Authentication.RateLimits.General.PerIP, nil)
	case RateLimitAuthenticationTOTP:
		return resolveConfig(cfg.Authentication.RateLimits.TOTP.PerIP, cfg.Authentication.RateLimits.General.PerIP, nil)
	case RateLimitAuthenticationRecoveryCode:
		return resolveConfig(cfg.Authentication.RateLimits.RecoveryCode.PerIP, cfg.Authentication.RateLimits.General.PerIP, nil)
	case RateLimitAuthenticationDeviceToken:
		return resolveConfig(cfg.Authentication.RateLimits.DeviceToken.PerIP, cfg.Authentication.RateLimits.General.PerIP, nil)
	case RateLimitAuthenticationPasskey:
		return resolveConfig(cfg.Authentication.RateLimits.Passkey.PerIP, cfg.Authentication.RateLimits.General.PerIP, nil)
	case RateLimitAuthenticationSIWE:
		return resolveConfig(cfg.Authentication.RateLimits.SIWE.PerIP, cfg.Authentication.RateLimits.General.PerIP, nil)
	case RateLimitAuthenticationSignup:
		return resolveConfig(cfg.Authentication.RateLimits.Signup.PerIP, nil, nil)
	case RateLimitAuthenticationSignupAnonymous:
		return resolveConfig(cfg.Authentication.RateLimits.SignupAnonymous.PerIP, nil, nil)
	case RateLimitAuthenticationAccountEnumeration:
		return resolveConfig(cfg.Authentication.RateLimits.AccountEnumeration.PerIP, nil, nil)

	// Features
	case RateLimitVerificationEmailTrigger:
		return resolveConfig(cfg.Verification.RateLimits.Email.TriggerPerIP, nil, nil)
	case RateLimitVerificationEmailValidate:
		return resolveConfig(cfg.Verification.RateLimits.Email.ValidatePerIP, nil, nil)
	case RateLimitVerificationSMSTrigger:
		return resolveConfig(cfg.Verification.RateLimits.SMS.TriggerPerIP, nil, nil)
	case RateLimitVerificationSMSValidate:
		return resolveConfig(cfg.Verification.RateLimits.SMS.ValidatePerIP, nil, nil)
	case RateLimitForgotPasswordEmailTrigger:
		return resolveConfig(cfg.ForgotPassword.RateLimits.Email.TriggerPerIP, nil, nil)
	case RateLimitForgotPasswordEmailValidate:
		return resolveConfig(cfg.ForgotPassword.RateLimits.Email.ValidatePerIP, nil, nil)
	case RateLimitForgotPasswordSMSTrigger:
		return resolveConfig(cfg.ForgotPassword.RateLimits.SMS.TriggerPerIP, nil, nil)
	case RateLimitForgotPasswordSMSValidate:
		return resolveConfig(cfg.ForgotPassword.RateLimits.SMS.ValidatePerIP, nil, nil)

	// Messaging
	case RateLimitMessagingSMS:
		return resolveConfig(cfg.Messaging.RateLimits.SMSPerIP, nil, featureCfg.Messaging.RateLimits.SMSPerIP)
	case RateLimitMessagingEmail:
		return resolveConfig(cfg.Messaging.RateLimits.EmailPerIP, nil, featureCfg.Messaging.RateLimits.EmailPerIP)

	}
	return nil
}

func (n RateLimitGroup) resolvePerUser(cfg *config.AppConfig) *config.RateLimitConfig {
	switch n {
	case RateLimitAuthenticationOOBOTPEmailTrigger:
		return resolveConfig(cfg.Authentication.RateLimits.OOBOTP.Email.TriggerPerUser, nil, nil)
	case RateLimitAuthenticationOOBOTPSMSTrigger:
		return resolveConfig(cfg.Authentication.RateLimits.OOBOTP.SMS.TriggerPerUser, nil, nil)
	case RateLimitVerificationEmailTrigger:
		return resolveConfig(cfg.Verification.RateLimits.Email.TriggerPerUser, nil, nil)
	case RateLimitVerificationSMSTrigger:
		return resolveConfig(cfg.Verification.RateLimits.SMS.TriggerPerUser, nil, nil)

	}
	return nil
}

func (n RateLimitGroup) resolvePerTarget(cfg *config.AppConfig, featureCfg *config.FeatureConfig) *config.RateLimitConfig {
	switch n {
	case RateLimitMessagingEmail:
		return resolveConfig(cfg.Messaging.RateLimits.EmailPerTarget, nil, featureCfg.Messaging.RateLimits.EmailPerTarget)
	case RateLimitMessagingSMS:
		return resolveConfig(cfg.Messaging.RateLimits.SMSPerTarget, nil, featureCfg.Messaging.RateLimits.SMSPerTarget)
	}
	return nil
}

func (n RateLimitGroup) resolvePerUserPerIP(cfg *config.AppConfig) *config.RateLimitConfig {
	switch n {
	case RateLimitAuthenticationGeneral:
		return resolveConfig(cfg.Authentication.RateLimits.General.PerUserPerIP, nil, nil)
	case RateLimitAuthenticationPassword:
		return resolveConfig(cfg.Authentication.RateLimits.Password.PerUserPerIP, cfg.Authentication.RateLimits.General.PerUserPerIP, nil)
	case RateLimitAuthenticationOOBOTPEmailValidate:
		return resolveConfig(cfg.Authentication.RateLimits.OOBOTP.Email.ValidatePerUserPerIP, cfg.Authentication.RateLimits.General.PerUserPerIP, nil)
	case RateLimitAuthenticationOOBOTPSMSValidate:
		return resolveConfig(cfg.Authentication.RateLimits.OOBOTP.SMS.ValidatePerUserPerIP, cfg.Authentication.RateLimits.General.PerUserPerIP, nil)
	case RateLimitAuthenticationTOTP:
		return resolveConfig(cfg.Authentication.RateLimits.TOTP.PerUserPerIP, cfg.Authentication.RateLimits.General.PerUserPerIP, nil)
	case RateLimitAuthenticationRecoveryCode:
		return resolveConfig(cfg.Authentication.RateLimits.RecoveryCode.PerUserPerIP, cfg.Authentication.RateLimits.General.PerUserPerIP, nil)
	case RateLimitAuthenticationDeviceToken:
		return resolveConfig(cfg.Authentication.RateLimits.DeviceToken.PerUserPerIP, cfg.Authentication.RateLimits.General.PerUserPerIP, nil)
	}
	return nil
}

func (n RateLimitGroup) resolvePerProject(cfg *config.AppConfig) *config.RateLimitConfig {
	switch n {
	case RateLimitOAuthTokenClientCredentials:
		return resolveConfig(&config.RateLimitConfig{
			Enabled: func() *bool { var t = true; return &t }(),
			Period:  "1m",
			Burst:   20,
		}, nil, nil)
	}
	return nil
}

func (n RateLimitGroup) resolvePerClient(cfg *config.AppConfig) *config.RateLimitConfig {
	switch n {
	case RateLimitOAuthTokenClientCredentials:
		return resolveConfig(&config.RateLimitConfig{
			Enabled: func() *bool { var t = true; return &t }(),
			Period:  "1m",
			Burst:   5,
		}, nil, nil)
	}
	return nil
}

type ResolveBucketSpecOptions struct {
	IPAddress string
	UserID    string
	Purpose   string
	Channel   model.AuthenticatorOOBChannel
	Target    string
	ClientID  string
}

func (r RateLimitGroup) ResolveBucketSpecs(
	cfg *config.AppConfig,
	featureCfg *config.FeatureConfig,
	globalCfg *config.RateLimitsEnvironmentConfig,
	opts *ResolveBucketSpecOptions,
) []*BucketSpec {
	var specs []*BucketSpec

	resolveBucket := func(rlCfg *config.RateLimitConfig, featureRlCfg *config.RateLimitConfig, bucketName BucketName, args ...string) *BucketSpec {
		effectiveCfg := rlCfg
		if featureRlCfg != nil && featureRlCfg.Rate() < rlCfg.Rate() {
			effectiveCfg = featureRlCfg
		}
		spec := NewBucketSpec(r, effectiveCfg, bucketName, args...)
		return &spec
	}

	resolvePerTargetBucket := func(bucketName BucketName, args ...string) *BucketSpec {
		confPerTarget := r.resolvePerTarget(cfg, featureCfg)
		if confPerTarget == nil {
			return &BucketSpecDisabled
		}
		spec := NewBucketSpec(r, confPerTarget, bucketName, args...)
		return &spec
	}

	resolvePerIPBucket := func(bucketName BucketName, args ...string) *BucketSpec {
		confPerIP := r.resolvePerIP(cfg, featureCfg)
		if confPerIP == nil {
			return &BucketSpecDisabled
		}
		spec := NewBucketSpec(r, confPerIP, bucketName, args...)
		return &spec
	}

	resolvePerUserBucket := func(bucketName BucketName, userID string, args ...string) *BucketSpec {
		confPerUser := r.resolvePerUser(cfg)
		if confPerUser == nil || userID == "" {
			return &BucketSpecDisabled
		}
		bucketArgs := []string{userID}
		bucketArgs = append(bucketArgs, args...)
		spec := NewBucketSpec(r, confPerUser, bucketName, bucketArgs...)
		return &spec
	}

	resolvePerUserPerIPBucket := func(bucketName BucketName, userID string, args ...string) *BucketSpec {
		confPerUserPerIP := r.resolvePerUserPerIP(cfg)
		if confPerUserPerIP == nil || userID == "" {
			return &BucketSpecDisabled
		}
		bucketArgs := []string{userID}
		bucketArgs = append(bucketArgs, args...)
		spec := NewBucketSpec(r, confPerUserPerIP, bucketName, bucketArgs...)
		return &spec
	}

	resolvePerClientBucket := func(bucketName BucketName, clientID string, args ...string) *BucketSpec {
		confPerClient := r.resolvePerClient(cfg)
		if confPerClient == nil || clientID == "" {
			return &BucketSpecDisabled
		}
		bucketArgs := []string{clientID}
		bucketArgs = append(bucketArgs, args...)
		spec := NewBucketSpec(r, confPerClient, bucketName, bucketArgs...)
		return &spec
	}

	resolvePerProjectBucket := func(bucketName BucketName, args ...string) *BucketSpec {
		confPerProject := r.resolvePerProject(cfg)
		if confPerProject == nil {
			return &BucketSpecDisabled
		}
		spec := NewBucketSpec(r, confPerProject, bucketName, args...)
		return &spec
	}

	resolveGlobalBucket := func(globalEntry config.RateLimitsEnvironmentConfigEntry, bucketName BucketName, args ...string) *BucketSpec {
		globalLimit := NewGlobalBucketSpec(r, globalEntry, bucketName, args...)
		return &globalLimit
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
		specs = append(specs, resolvePerUserBucket(bucketNamePerUser, opts.UserID, opts.Purpose))

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
		specs = append(specs, resolvePerUserPerIPBucket(bucketNamePerUserPerIP, opts.UserID, opts.IPAddress, opts.Purpose))

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
		specs = append(specs, resolvePerUserBucket(bucketNamePerUser, opts.UserID, opts.Purpose))

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
		specs = append(specs, resolvePerUserPerIPBucket(bucketNamePerUserPerIP, opts.UserID, opts.IPAddress, opts.Purpose))

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
		bucketName := selectByChannel(opts.Channel,
			ForgotPasswordTriggerEmailPerIP,
			ForgotPasswordTriggerSMSPerIP,
			ForgotPasswordTriggerWhatsappPerIP,
		)
		specs = append(specs, resolvePerIPBucket(bucketName, opts.IPAddress))

	case RateLimitForgotPasswordEmailValidate:
		bucketName := selectByChannel(opts.Channel,
			ForgotPasswordValidateEmailPerIP,
			ForgotPasswordValidateSMSPerIP,
			ForgotPasswordValidateWhatsappPerIP,
		)
		specs = append(specs, resolvePerIPBucket(bucketName, opts.IPAddress))

	case RateLimitForgotPasswordSMSTrigger:
		bucketName := selectByChannel(opts.Channel,
			ForgotPasswordTriggerEmailPerIP,
			ForgotPasswordTriggerSMSPerIP,
			ForgotPasswordTriggerWhatsappPerIP,
		)
		specs = append(specs, resolvePerIPBucket(bucketName, opts.IPAddress))

	case RateLimitForgotPasswordSMSValidate:
		bucketName := selectByChannel(opts.Channel,
			ForgotPasswordValidateEmailPerIP,
			ForgotPasswordValidateSMSPerIP,
			ForgotPasswordValidateWhatsappPerIP,
		)
		specs = append(specs, resolvePerIPBucket(bucketName, opts.IPAddress))

	case RateLimitMessagingSMS:
		specs = append(specs, resolveBucket(cfg.Messaging.RateLimits.SMS, featureCfg.Messaging.RateLimits.SMS, MessagingSMS))
		specs = append(specs, resolveGlobalBucket(globalCfg.SMS, MessagingSMS))
		specs = append(specs, resolvePerIPBucket(MessagingSMSPerIP, opts.IPAddress))
		specs = append(specs, resolveGlobalBucket(globalCfg.SMSPerIP, MessagingSMSPerIP, opts.IPAddress))
		specs = append(specs, resolvePerTargetBucket(MessagingSMSPerTarget, opts.Target))
		specs = append(specs, resolveGlobalBucket(globalCfg.SMSPerTarget, MessagingSMSPerTarget, opts.Target))

	case RateLimitMessagingEmail:
		specs = append(specs, resolveBucket(cfg.Messaging.RateLimits.Email, featureCfg.Messaging.RateLimits.Email, MessagingEmail))
		specs = append(specs, resolveGlobalBucket(globalCfg.Email, MessagingEmail))
		specs = append(specs, resolvePerIPBucket(MessagingEmailPerIP, opts.IPAddress))
		specs = append(specs, resolveGlobalBucket(globalCfg.EmailPerIP, MessagingEmailPerIP, opts.IPAddress))
		specs = append(specs, resolvePerTargetBucket(MessagingEmailPerTarget, opts.Target))
		specs = append(specs, resolveGlobalBucket(globalCfg.EmailPerTarget, MessagingEmailPerTarget, opts.Target))

	case RateLimitOAuthTokenClientCredentials:
		specs = append(specs, resolvePerClientBucket(OAuthTokenClientCredentialsPerClient, opts.ClientID))
		specs = append(specs, resolvePerProjectBucket(OAuthTokenClientCredentialsPerProject))

	case RateLimitOAuthTokenGeneral:
		panic(fmt.Errorf("ResolveBucketSpecs not supported for %s", RateLimitOAuthTokenGeneral))
	}

	return specs
}

func (r RateLimitGroup) ResolveWeight(
	ctx context.Context,
) float64 {
	var defaultWeight float64 = 1
	weights := getRateLimitWeights(ctx)
	if weights == nil {
		return defaultWeight
	}
	resolveWeight := func(selfWeightKey RateLimitGroup, fallbackKey RateLimitGroup) float64 {
		if selfWeight, ok := weights[selfWeightKey]; ok {
			return selfWeight
		}

		if fallbackKey != "" {
			if fallbackWeight, ok := weights[fallbackKey]; ok {
				return fallbackWeight
			}
		}

		return defaultWeight
	}

	var weight float64
	switch r {
	case "":
		// Handle unspecified rate limits
		return defaultWeight
	case RateLimitAuthenticationPassword,
		RateLimitAuthenticationOOBOTPEmailValidate,
		RateLimitAuthenticationOOBOTPSMSValidate,
		RateLimitAuthenticationTOTP,
		RateLimitAuthenticationRecoveryCode,
		RateLimitAuthenticationDeviceToken,
		RateLimitAuthenticationPasskey,
		RateLimitAuthenticationSIWE:
		weight = resolveWeight(r, RateLimitAuthenticationGeneral)
	default:
		weight = resolveWeight(r, "")
	}

	if weight < 0 {
		weight = 0
	}
	// Weight can never be < 0
	return weight
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

func resolveConfig(c *config.RateLimitConfig, fallback *config.RateLimitConfig, featureConfig *config.RateLimitConfig) *config.RateLimitConfig {
	effectiveConfig := c

	// We check Enabled == nil here so that only unspecified rate limits are fallbacked
	// If it is enabled or disabled explicitly, fallback is not used
	if effectiveConfig.Enabled == nil && fallback != nil {
		effectiveConfig = fallback
	}

	if featureConfig != nil && featureConfig.Rate() < c.Rate() {
		effectiveConfig = featureConfig
	}

	return effectiveConfig
}
