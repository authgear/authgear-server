package config

var _ = Schema.Add("AuthenticationRateLimitsConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"general": { "$ref": "#/$defs/AuthenticationRateLimitsGeneralConfig" },
		"password": { "$ref": "#/$defs/AuthenticationRateLimitsPasswordConfig" },
		"oob_otp": { "$ref": "#/$defs/AuthenticationRateLimitsOOBOTPConfig" },
		"totp": { "$ref": "#/$defs/AuthenticationRateLimitsTOTPConfig" },
		"passkey": { "$ref": "#/$defs/AuthenticationRateLimitsPasskeyConfig" },
		"siwe": { "$ref": "#/$defs/AuthenticationRateLimitsSIWEConfig" },
		"recovery_code": { "$ref": "#/$defs/AuthenticationRateLimitsRecoveryCodeConfig" },
		"device_token": { "$ref": "#/$defs/AuthenticationRateLimitsDeviceTokenConfig" },
		"signup": { "$ref": "#/$defs/AuthenticationRateLimitsSignupConfig" },
		"signup_anonymous": { "$ref": "#/$defs/AuthenticationRateLimitsSignupAnonymousConfig" },
		"account_enumeration": { "$ref": "#/$defs/AuthenticationRateLimitsAccountEnumerationConfig" }
	}
}
`)

type AuthenticationRateLimitsConfig struct {
	General      *AuthenticationRateLimitsGeneralConfig      `json:"general,omitempty"`
	Password     *AuthenticationRateLimitsPasswordConfig     `json:"password,omitempty"`
	OOBOTP       *AuthenticationRateLimitsOOBOTPConfig       `json:"oob_otp,omitempty"`
	TOTP         *AuthenticationRateLimitsTOTPConfig         `json:"totp,omitempty"`
	Passkey      *AuthenticationRateLimitsPasskeyConfig      `json:"passkey,omitempty"`
	SIWE         *AuthenticationRateLimitsSIWEConfig         `json:"siwe,omitempty"`
	RecoveryCode *AuthenticationRateLimitsRecoveryCodeConfig `json:"recovery_code,omitempty"`
	DeviceToken  *AuthenticationRateLimitsDeviceTokenConfig  `json:"device_token,omitempty"`

	Signup             *AuthenticationRateLimitsSignupConfig             `json:"signup,omitempty"`
	SignupAnonymous    *AuthenticationRateLimitsSignupAnonymousConfig    `json:"signup_anonymous,omitempty"`
	AccountEnumeration *AuthenticationRateLimitsAccountEnumerationConfig `json:"account_enumeration,omitempty"`
}

var _ = Schema.Add("AuthenticationRateLimitsGeneralConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"per_ip": { "$ref": "#/$defs/RateLimitConfig" },
		"per_user_per_ip": { "$ref": "#/$defs/RateLimitConfig" }
	}
}
`)

type AuthenticationRateLimitsGeneralConfig struct {
	PerIP        *RateLimitConfig `json:"per_ip,omitempty"`
	PerUserPerIP *RateLimitConfig `json:"per_user_per_ip,omitempty"`
}

func (c *AuthenticationRateLimitsGeneralConfig) SetDefaults() {
	if c.PerIP.Enabled == nil {
		c.PerIP = &RateLimitConfig{
			Enabled: newBool(true),
			Period:  "1m",
			Burst:   60,
		}
	}
	if c.PerUserPerIP.Enabled == nil {
		c.PerUserPerIP = &RateLimitConfig{
			Enabled: newBool(true),
			Period:  "1m",
			Burst:   10,
		}
	}
}

var _ = Schema.Add("AuthenticationRateLimitsPasswordConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"per_ip": { "$ref": "#/$defs/RateLimitConfig" },
		"per_user_per_ip": { "$ref": "#/$defs/RateLimitConfig" }
	}
}
`)

type AuthenticationRateLimitsPasswordConfig struct {
	PerIP        *RateLimitConfig `json:"per_ip,omitempty"`
	PerUserPerIP *RateLimitConfig `json:"per_user_per_ip,omitempty"`
}

var _ = Schema.Add("AuthenticationRateLimitsOOBOTPConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"email": { "$ref": "#/$defs/AuthenticationRateLimitsOOBOTPEmailConfig" },
		"sms": { "$ref": "#/$defs/AuthenticationRateLimitsOOBOTPSMSConfig" }
	}
}
`)

type AuthenticationRateLimitsOOBOTPConfig struct {
	Email *AuthenticationRateLimitsOOBOTPEmailConfig `json:"email,omitempty"`
	SMS   *AuthenticationRateLimitsOOBOTPSMSConfig   `json:"sms,omitempty"`
}

var _ = Schema.Add("AuthenticationRateLimitsOOBOTPEmailConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"trigger_per_ip": { "$ref": "#/$defs/RateLimitConfig" },
		"trigger_per_user": { "$ref": "#/$defs/RateLimitConfig" },
		"trigger_cooldown": { "$ref": "#/$defs/DurationString" },
		"max_failed_attempts_revoke_otp": { "type": "integer", "minimum": 1 },
		"validate_per_ip": { "$ref": "#/$defs/RateLimitConfig" },
		"validate_per_user_per_ip": { "$ref": "#/$defs/RateLimitConfig" }
	}
}
`)

type AuthenticationRateLimitsOOBOTPEmailConfig struct {
	TriggerPerIP               *RateLimitConfig `json:"trigger_per_ip,omitempty"`
	TriggerPerUser             *RateLimitConfig `json:"trigger_per_user,omitempty"`
	TriggerCooldown            DurationString   `json:"trigger_cooldown,omitempty"`
	MaxFailedAttemptsRevokeOTP int              `json:"max_failed_attempts_revoke_otp,omitempty"`
	ValidatePerIP              *RateLimitConfig `json:"validate_per_ip,omitempty"`
	ValidatePerUserPerIP       *RateLimitConfig `json:"validate_per_user_per_ip,omitempty"`
}

func (c *AuthenticationRateLimitsOOBOTPEmailConfig) SetDefaults() {
	if c.TriggerCooldown == "" {
		c.TriggerCooldown = "1m"
	}
}

var _ = Schema.Add("AuthenticationRateLimitsOOBOTPSMSConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"trigger_per_ip": { "$ref": "#/$defs/RateLimitConfig" },
		"trigger_per_user": { "$ref": "#/$defs/RateLimitConfig" },
		"trigger_cooldown": { "$ref": "#/$defs/DurationString" },
		"max_failed_attempts_revoke_otp": { "type": "integer", "minimum": 1 },
		"validate_per_ip": { "$ref": "#/$defs/RateLimitConfig" },
		"validate_per_user_per_ip": { "$ref": "#/$defs/RateLimitConfig" }
	}
}
`)

type AuthenticationRateLimitsOOBOTPSMSConfig struct {
	TriggerPerIP               *RateLimitConfig `json:"trigger_per_ip,omitempty"`
	TriggerPerUser             *RateLimitConfig `json:"trigger_per_user,omitempty"`
	TriggerCooldown            DurationString   `json:"trigger_cooldown,omitempty"`
	MaxFailedAttemptsRevokeOTP int              `json:"max_failed_attempts_revoke_otp,omitempty"`
	ValidatePerIP              *RateLimitConfig `json:"validate_per_ip,omitempty"`
	ValidatePerUserPerIP       *RateLimitConfig `json:"validate_per_user_per_ip,omitempty"`
}

func (c *AuthenticationRateLimitsOOBOTPSMSConfig) SetDefaults() {
	if c.TriggerCooldown == "" {
		c.TriggerCooldown = "1m"
	}
}

var _ = Schema.Add("AuthenticationRateLimitsTOTPConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"per_ip": { "$ref": "#/$defs/RateLimitConfig" },
		"per_user_per_ip": { "$ref": "#/$defs/RateLimitConfig" }
	}
}
`)

type AuthenticationRateLimitsTOTPConfig struct {
	PerIP        *RateLimitConfig `json:"per_ip,omitempty"`
	PerUserPerIP *RateLimitConfig `json:"per_user_per_ip,omitempty"`
}

var _ = Schema.Add("AuthenticationRateLimitsPasskeyConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"per_ip": { "$ref": "#/$defs/RateLimitConfig" }
	}
}
`)

type AuthenticationRateLimitsPasskeyConfig struct {
	PerIP *RateLimitConfig `json:"per_ip,omitempty"`
}

var _ = Schema.Add("AuthenticationRateLimitsSIWEConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"per_ip": { "$ref": "#/$defs/RateLimitConfig" }
	}
}
`)

type AuthenticationRateLimitsSIWEConfig struct {
	PerIP *RateLimitConfig `json:"per_ip,omitempty"`
}

var _ = Schema.Add("AuthenticationRateLimitsRecoveryCodeConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"per_ip": { "$ref": "#/$defs/RateLimitConfig" },
		"per_user_per_ip": { "$ref": "#/$defs/RateLimitConfig" }
	}
}
`)

type AuthenticationRateLimitsRecoveryCodeConfig struct {
	PerIP        *RateLimitConfig `json:"per_ip,omitempty"`
	PerUserPerIP *RateLimitConfig `json:"per_user_per_ip,omitempty"`
}

var _ = Schema.Add("AuthenticationRateLimitsDeviceTokenConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"per_ip": { "$ref": "#/$defs/RateLimitConfig" },
		"per_user_per_ip": { "$ref": "#/$defs/RateLimitConfig" }
	}
}
`)

type AuthenticationRateLimitsDeviceTokenConfig struct {
	PerIP        *RateLimitConfig `json:"per_ip,omitempty"`
	PerUserPerIP *RateLimitConfig `json:"per_user_per_ip,omitempty"`
}

var _ = Schema.Add("AuthenticationRateLimitsSignupConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"per_ip": { "$ref": "#/$defs/RateLimitConfig" }
	}
}
`)

type AuthenticationRateLimitsSignupConfig struct {
	PerIP *RateLimitConfig `json:"per_ip,omitempty"`
}

func (c *AuthenticationRateLimitsSignupConfig) SetDefaults() {
	if c.PerIP.Enabled == nil {
		c.PerIP = &RateLimitConfig{
			Enabled: newBool(true),
			Period:  "1m",
			Burst:   10,
		}
	}
}

var _ = Schema.Add("AuthenticationRateLimitsSignupAnonymousConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"per_ip": { "$ref": "#/$defs/RateLimitConfig" }
	}
}
`)

type AuthenticationRateLimitsSignupAnonymousConfig struct {
	PerIP *RateLimitConfig `json:"per_ip,omitempty"`
}

func (c *AuthenticationRateLimitsSignupAnonymousConfig) SetDefaults() {
	if c.PerIP.Enabled == nil {
		c.PerIP = &RateLimitConfig{
			Enabled: newBool(true),
			Period:  "1m",
			Burst:   60,
		}
	}
}

var _ = Schema.Add("AuthenticationRateLimitsAccountEnumerationConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"per_ip": { "$ref": "#/$defs/RateLimitConfig" }
	}
}
`)

type AuthenticationRateLimitsAccountEnumerationConfig struct {
	PerIP *RateLimitConfig `json:"per_ip,omitempty"`
}

func (c *AuthenticationRateLimitsAccountEnumerationConfig) SetDefaults() {
	if c.PerIP.Enabled == nil {
		c.PerIP = &RateLimitConfig{
			Enabled: newBool(true),
			Period:  "1m",
			Burst:   10,
		}
	}
}
