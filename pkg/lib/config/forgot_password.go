package config

var _ = Schema.Add("ForgotPasswordConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"enabled": { "type": "boolean" },
		"reset_code_expiry_seconds": { "$ref": "#/$defs/DurationSeconds" },
		"code_valid_period": { "$ref": "#/$defs/DurationString" },
		"rate_limits": { "$ref": "#/$defs/ForgotPasswordRateLimitsConfig" }
	}
}
`)

type ForgotPasswordConfig struct {
	Enabled *bool `json:"enabled,omitempty"`

	// ResetCodeExpiry is deprecated
	ResetCodeExpiry DurationSeconds `json:"reset_code_expiry_seconds,omitempty"`
	// CodeValidPeriod is deprecated
	CodeValidPeriod DurationString `json:"code_valid_period,omitempty"`
	LinkValidPeriod DurationString `json:"link_valid_period,omitempty"`
	OTPValidPeriod  DurationString `json:"otp_valid_period,omitempty"`

	RateLimits *ForgotPasswordRateLimitsConfig `json:"rate_limits,omitempty"`
}

func (c *ForgotPasswordConfig) SetDefaults() {
	if c.Enabled == nil {
		c.Enabled = newBool(true)
	}

	if c.ResetCodeExpiry == 0 {
		// https://cheatsheetseries.owasp.org/cheatsheets/Forgot_Password_Cheat_Sheet.html#step-3-send-a-token-over-a-side-channel
		// OWASP suggests the lifetime is no more than 20 minutes
		c.ResetCodeExpiry = DurationSeconds(1200)
	}
	if c.CodeValidPeriod == "" {
		c.CodeValidPeriod = DurationString(c.ResetCodeExpiry.Duration().String())
	}
	if c.LinkValidPeriod == "" {
		c.LinkValidPeriod = c.CodeValidPeriod
	}

	if c.OTPValidPeriod == "" {
		c.OTPValidPeriod = DurationString("300s")
	}
}

var _ = Schema.Add("ForgotPasswordRateLimitsConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"email": { "$ref": "#/$defs/ForgotPasswordRateLimitsEmailConfig" },
		"sms": { "$ref": "#/$defs/ForgotPasswordRateLimitsSMSConfig" }
	}
}
`)

type ForgotPasswordRateLimitsConfig struct {
	Email *ForgotPasswordRateLimitsEmailConfig `json:"email,omitempty"`
	SMS   *ForgotPasswordRateLimitsSMSConfig   `json:"sms,omitempty"`
}

var _ = Schema.Add("ForgotPasswordRateLimitsEmailConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"trigger_per_ip": { "$ref": "#/$defs/RateLimitConfig" },
		"trigger_cooldown": { "$ref": "#/$defs/DurationString" },
		"validate_per_ip": { "$ref": "#/$defs/RateLimitConfig" }
	}
}
`)

type ForgotPasswordRateLimitsEmailConfig struct {
	TriggerPerIP    *RateLimitConfig `json:"trigger_per_ip,omitempty"`
	TriggerCooldown DurationString   `json:"trigger_cooldown,omitempty"`
	ValidatePerIP   *RateLimitConfig `json:"validate_per_ip,omitempty"`
}

func (c *ForgotPasswordRateLimitsEmailConfig) SetDefaults() {
	if c.TriggerCooldown == "" {
		c.TriggerCooldown = "1m"
	}
	if c.ValidatePerIP.Enabled == nil {
		c.ValidatePerIP = &RateLimitConfig{
			Enabled: newBool(true),
			Period:  "1m",
			Burst:   60,
		}
	}
}

var _ = Schema.Add("ForgotPasswordRateLimitsSMSConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"trigger_per_ip": { "$ref": "#/$defs/RateLimitConfig" },
		"trigger_cooldown": { "$ref": "#/$defs/DurationString" },
		"validate_per_ip": { "$ref": "#/$defs/RateLimitConfig" }
	}
}
`)

type ForgotPasswordRateLimitsSMSConfig struct {
	TriggerPerIP    *RateLimitConfig `json:"trigger_per_ip,omitempty"`
	TriggerCooldown DurationString   `json:"trigger_cooldown,omitempty"`
	ValidatePerIP   *RateLimitConfig `json:"validate_per_ip,omitempty"`
}

func (c *ForgotPasswordRateLimitsSMSConfig) SetDefaults() {
	if c.TriggerCooldown == "" {
		c.TriggerCooldown = "1m"
	}
	if c.ValidatePerIP.Enabled == nil {
		c.ValidatePerIP = &RateLimitConfig{
			Enabled: newBool(true),
			Period:  "1m",
			Burst:   60,
		}
	}
}
