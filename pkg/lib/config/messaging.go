package config

var _ = Schema.Add("MessagingConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"sms_provider": { "$ref": "#/$defs/SMSProvider" },
		"sms_gateway": { "$ref": "#/$defs/SMSGatewayConfig" },
		"sms": { "$ref": "#/$defs/SMSConfig" },
		"email": { "$ref": "#/$defs/EmailConfig" },
		"whatsapp": { "$ref": "#/$defs/WhatsappConfig" },
		"rate_limits": { "$ref": "#/$defs/MessagingRateLimitsConfig" }
	}
}
`)

type MessagingConfig struct {
	SMSProvider      SMSProvider                `json:"sms_provider,omitempty"`
	SMSGateway       *SMSGatewayConfig          `json:"sms_gateway,omitempty" nullable:"true"`
	Deprecated_SMS   *SMSConfig                 `json:"sms,omitempty"`
	Deprecated_Email *EmailConfig               `json:"email,omitempty"`
	Whatsapp         *WhatsappConfig            `json:"whatsapp,omitempty"`
	RateLimits       *MessagingRateLimitsConfig `json:"rate_limits,omitempty"`
}

func (c *MessagingConfig) SetDefaults() {
	c.Deprecated_SMS = nil
	c.Deprecated_Email = nil
}

var _ = Schema.Add("SMSProvider", `
{
	"type": "string",
	"enum": ["nexmo", "twilio", "custom"]
}
`)

type SMSProvider string

const (
	SMSProviderNexmo  SMSProvider = "nexmo"
	SMSProviderTwilio SMSProvider = "twilio"
	SMSProviderCustom SMSProvider = "custom"
)

var _ = Schema.Add("SMSConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"ratelimit": { "$ref": "#/$defs/SMSRatelimitConfig" }
	}
}
`)

// SMSConfig is deprecated.
type SMSConfig struct {
	Ratelimit *SMSRatelimitConfig `json:"ratelimit,omitempty"`
}

var _ = Schema.Add("SMSRatelimitConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"per_phone": { "$ref": "#/$defs/SMSRateLimitPerPhoneConfig" },
		"per_ip": { "$ref": "#/$defs/SMSRateLimitPerIPConfig" },
		"resend_cooldown_seconds": {
			"$ref": "#/$defs/DurationSeconds",
			"enum": [60, 120]
		}
	}
}
`)

type SMSRatelimitConfig struct {
	PerPhone              *SMSRateLimitPerPhoneConfig `json:"per_phone,omitempty"`
	PerIP                 *SMSRateLimitPerIPConfig    `json:"per_ip,omitempty"`
	ResendCooldownSeconds DurationSeconds             `json:"resend_cooldown_seconds,omitempty"`
}

func (c *SMSRatelimitConfig) SetDefaults() {
	if c.ResendCooldownSeconds == 0 {
		c.ResendCooldownSeconds = DurationSeconds(60)
	}
}

var _ = Schema.Add("SMSRateLimitPerPhoneConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"enabled": { "type": "boolean" },
		"size": { "type": "integer", "minimum": 1, "maximum": 100 },
		"reset_period": { "$ref": "#/$defs/DurationString" }
	}
}
`)

type SMSRateLimitPerPhoneConfig struct {
	Enabled     bool           `json:"enabled,omitempty"`
	Size        int            `json:"size,omitempty"`
	ResetPeriod DurationString `json:"reset_period,omitempty"`
}

func (c *SMSRateLimitPerPhoneConfig) SetDefaults() {
	if c.Enabled {
		if c.Size == 0 {
			c.Size = 10
		}
		if c.ResetPeriod == "" {
			c.ResetPeriod = "24h"
		}
	}
}

var _ = Schema.Add("SMSRateLimitPerIPConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"enabled": { "type": "boolean" },
		"size": { "type": "integer", "minimum": 1 },
		"reset_period": { "$ref": "#/$defs/DurationString" }
	}
}
`)

type SMSRateLimitPerIPConfig struct {
	Enabled     bool           `json:"enabled,omitempty"`
	Size        int            `json:"size,omitempty"`
	ResetPeriod DurationString `json:"reset_period,omitempty"`
}

func (c *SMSRateLimitPerIPConfig) SetDefaults() {
	if c.Enabled {
		if c.Size == 0 {
			c.Size = 120
		}
		if c.ResetPeriod == "" {
			c.ResetPeriod = "1m"
		}
	}
}

var _ = Schema.Add("EmailConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"ratelimit": { "$ref": "#/$defs/EmailRatelimitConfig" }
	}
}
`)

// EmailConfig is deprecated.
type EmailConfig struct {
	Ratelimit *EmailRatelimitConfig `json:"ratelimit,omitempty"`
}

var _ = Schema.Add("EmailRatelimitConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"resend_cooldown_seconds": {
			"$ref": "#/$defs/DurationSeconds",
			"enum": [60, 120]
		}
	}
}
`)

type EmailRatelimitConfig struct {
	ResendCooldownSeconds DurationSeconds `json:"resend_cooldown_seconds,omitempty"`
}

func (c *EmailRatelimitConfig) SetDefaults() {
	if c.ResendCooldownSeconds == 0 {
		c.ResendCooldownSeconds = DurationSeconds(60)
	}
}

var _ = Schema.Add("MessagingRateLimitsConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"sms": { "$ref": "#/$defs/RateLimitConfig" },
		"sms_per_ip": { "$ref": "#/$defs/RateLimitConfig" },
		"sms_per_target": { "$ref": "#/$defs/RateLimitConfig" },
		"email": { "$ref": "#/$defs/RateLimitConfig" },
		"email_per_ip": { "$ref": "#/$defs/RateLimitConfig" },
		"email_per_target": { "$ref": "#/$defs/RateLimitConfig" }
	}
}
`)

type MessagingRateLimitsConfig struct {
	SMS            *RateLimitConfig `json:"sms,omitempty"`
	SMSPerIP       *RateLimitConfig `json:"sms_per_ip,omitempty"`
	SMSPerTarget   *RateLimitConfig `json:"sms_per_target,omitempty"`
	Email          *RateLimitConfig `json:"email,omitempty"`
	EmailPerIP     *RateLimitConfig `json:"email_per_ip,omitempty"`
	EmailPerTarget *RateLimitConfig `json:"email_per_target,omitempty"`
}

func (c *MessagingRateLimitsConfig) SetDefaults() {
	if c.SMSPerIP.Enabled == nil {
		c.SMSPerIP = &RateLimitConfig{
			Enabled: newBool(true),
			Period:  "1m",
			Burst:   60,
		}
	}
	if c.SMSPerTarget.Enabled == nil {
		c.SMSPerTarget = &RateLimitConfig{
			Enabled: newBool(true),
			Period:  "1h",
			Burst:   10,
		}
	}
	if c.EmailPerIP.Enabled == nil {
		c.EmailPerIP = &RateLimitConfig{
			Enabled: newBool(true),
			Period:  "1m",
			Burst:   200,
		}
	}
	if c.EmailPerTarget.Enabled == nil {
		c.EmailPerTarget = &RateLimitConfig{
			Enabled: newBool(true),
			Period:  "24h",
			Burst:   50,
		}
	}
}
