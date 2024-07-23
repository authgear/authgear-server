package config

var _ = Schema.Add("VerificationCriteria", `
{
	"type": "string",
	"enum": ["any", "all"]
}
`)

type VerificationCriteria string

const (
	VerificationCriteriaAny VerificationCriteria = "any"
	VerificationCriteriaAll VerificationCriteria = "all"
)

var _ = Schema.Add("VerificationConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"claims": { "$ref": "#/$defs/VerificationClaimsConfig" },
		"criteria": { "$ref": "#/$defs/VerificationCriteria" },
		"rate_limits": { "$ref": "#/$defs/VerificationRateLimitsConfig" },
		"code_expiry_seconds": { "$ref": "#/$defs/DurationSeconds", "minimum": 60 },
		"code_valid_period": { "$ref": "#/$defs/DurationString" }
	}
}
`)

type VerificationConfig struct {
	Claims     *VerificationClaimsConfig     `json:"claims,omitempty"`
	Criteria   VerificationCriteria          `json:"criteria,omitempty"`
	RateLimits *VerificationRateLimitsConfig `json:"rate_limits,omitempty"`

	Deprecated_CodeExpirySeconds DurationSeconds `json:"code_expiry_seconds,omitempty"`
	CodeValidPeriod              DurationString  `json:"code_valid_period,omitempty"`
}

func (c *VerificationConfig) SetDefaults() {
	if c.Criteria == "" {
		c.Criteria = VerificationCriteriaAny
	}
	if c.Deprecated_CodeExpirySeconds == 0 {
		c.Deprecated_CodeExpirySeconds = DurationSeconds(300)
	}
	if c.CodeValidPeriod == "" {
		c.CodeValidPeriod = DurationString(c.Deprecated_CodeExpirySeconds.Duration().String())
	}

	c.Deprecated_CodeExpirySeconds = 0
}

var _ = Schema.Add("VerificationClaimsConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"email": { "$ref": "#/$defs/VerificationClaimConfig" },
		"phone_number": { "$ref": "#/$defs/VerificationClaimConfig" }
	}
}
`)

type VerificationClaimsConfig struct {
	Email       *VerificationClaimConfig `json:"email,omitempty"`
	PhoneNumber *VerificationClaimConfig `json:"phone_number,omitempty"`
}

var _ = Schema.Add("VerificationClaimConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"enabled": { "type": "boolean" },
		"required": { "type": "boolean" }
	}
}
`)

type VerificationClaimConfig struct {
	Enabled  *bool `json:"enabled,omitempty"`
	Required *bool `json:"required,omitempty"`
}

func (c *VerificationClaimConfig) SetDefaults() {
	if c.Enabled == nil {
		c.Enabled = newBool(true)
	}
	if c.Required == nil {
		c.Required = newBool(true)
	}
}

var _ = Schema.Add("VerificationRateLimitsConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"email": { "$ref": "#/$defs/VerificationRateLimitsEmailConfig" },
		"sms": { "$ref": "#/$defs/VerificationRateLimitsSMSConfig" }
	}
}
`)

type VerificationRateLimitsConfig struct {
	Email *VerificationRateLimitsEmailConfig `json:"email,omitempty"`
	SMS   *VerificationRateLimitsSMSConfig   `json:"sms,omitempty"`
}

var _ = Schema.Add("VerificationRateLimitsEmailConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"trigger_per_ip": { "$ref": "#/$defs/RateLimitConfig" },
		"trigger_per_user": { "$ref": "#/$defs/RateLimitConfig" },
		"trigger_cooldown": { "$ref": "#/$defs/DurationString" },
		"max_failed_attempts_revoke_otp": { "type": "integer", "minimum": 1 },
		"validate_per_ip": { "$ref": "#/$defs/RateLimitConfig" }
	}
}
`)

type VerificationRateLimitsEmailConfig struct {
	TriggerPerIP               *RateLimitConfig `json:"trigger_per_ip,omitempty"`
	TriggerPerUser             *RateLimitConfig `json:"trigger_per_user,omitempty"`
	TriggerCooldown            DurationString   `json:"trigger_cooldown,omitempty"`
	MaxFailedAttemptsRevokeOTP int              `json:"max_failed_attempts_revoke_otp,omitempty"`
	ValidatePerIP              *RateLimitConfig `json:"validate_per_ip,omitempty"`
}

func (c *VerificationRateLimitsEmailConfig) SetDefaults() {
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

var _ = Schema.Add("VerificationRateLimitsSMSConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"trigger_per_ip": { "$ref": "#/$defs/RateLimitConfig" },
		"trigger_per_user": { "$ref": "#/$defs/RateLimitConfig" },
		"trigger_cooldown": { "$ref": "#/$defs/DurationString" },
		"max_failed_attempts_revoke_otp": { "type": "integer", "minimum": 1 },
		"validate_per_ip": { "$ref": "#/$defs/RateLimitConfig" }
	}
}
`)

type VerificationRateLimitsSMSConfig struct {
	TriggerPerIP               *RateLimitConfig `json:"trigger_per_ip,omitempty"`
	TriggerPerUser             *RateLimitConfig `json:"trigger_per_user,omitempty"`
	TriggerCooldown            DurationString   `json:"trigger_cooldown,omitempty"`
	MaxFailedAttemptsRevokeOTP int              `json:"max_failed_attempts_revoke_otp,omitempty"`
	ValidatePerIP              *RateLimitConfig `json:"validate_per_ip,omitempty"`
}

func (c *VerificationRateLimitsSMSConfig) SetDefaults() {
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
