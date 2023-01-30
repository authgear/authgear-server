package config

var _ = Schema.Add("MessagingConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"sms_provider": { "$ref": "#/$defs/SMSProvider" },
		"sms": { "$ref": "#/$defs/SMSConfig" },
		"email": { "$ref": "#/$defs/EmailConfig" }
	}
}
`)

var _ = Schema.Add("CustomSMSProviderConfigs", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"url": { "type": "string" },
		"timeout": { "type": "integer" }
	},
	"required": ["url"]
}
`)

type CustomSMSProviderConfigs struct {
	URL     string `json:"url,omitempty"`
	Timeout int    `json:"timeout,omitempty"`
}

type MessagingConfig struct {
	SMSProvider SMSProvider  `json:"sms_provider,omitempty"`
	SMS         *SMSConfig   `json:"sms,omitempty"`
	Email       *EmailConfig `json:"email,omitempty"`
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

type SMSConfig struct {
	Ratelimit *SMSRatelimitConfig `json:"ratelimit,omitempty"`
}

var _ = Schema.Add("SMSRatelimitConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"per_phone": { "$ref": "#/$defs/SMSRateLimitPerPhoneConfig" },
		"resend_cooldown_seconds": {
			"$ref": "#/$defs/DurationSeconds",
			"enum": [60, 120]
		}
	}
}
`)

type SMSRatelimitConfig struct {
	PerPhone              *SMSRateLimitPerPhoneConfig `json:"per_phone,omitempty"`
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

var _ = Schema.Add("EmailConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"ratelimit": { "$ref": "#/$defs/EmailRatelimitConfig" }
	}
}
`)

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
