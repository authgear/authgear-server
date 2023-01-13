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

type MessagingConfig struct {
	SMSProvider SMSProvider  `json:"sms_provider,omitempty"`
	SMS         *SMSConfig   `json:"sms,omitempty"`
	Email       *EmailConfig `json:"email,omitempty"`
}

var _ = Schema.Add("SMSProvider", `
{
	"type": "string",
	"enum": ["nexmo", "twilio"]
}
`)

type SMSProvider string

const (
	SMSProviderNexmo  SMSProvider = "nexmo"
	SMSProviderTwilio SMSProvider = "twilio"
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
		"per_phone": { "$ref": "#/$defs/SMSPerPhoneConfig" },
		"resend_cooldown_seconds": {
			"$ref": "#/$defs/DurationSeconds",
			"enum": [60, 120]
		}
	}
}
`)

type SMSRatelimitConfig struct {
	PerPhone              SMSPerPhoneConfig `json:"per_phone,omitempty"`
	ResendCooldownSeconds DurationSeconds   `json:"resend_cooldown_seconds,omitempty"`
}

func (c *SMSRatelimitConfig) SetDefaults() {
	if c.ResendCooldownSeconds == 0 {
		c.ResendCooldownSeconds = DurationSeconds(60)
	}
}

var _ = Schema.Add("SMSPerPhoneConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"size": { "type": "integer", "minimum": 1, "maximum": 100 },
		"reset_period": { "$ref": "#/$defs/DurationString" }
	}
}
`)

type SMSPerPhoneConfig struct {
	Size        int            `json:"size,omitempty"`
	ResetPeriod DurationString `json:"reset_period,omitempty"`
}

func (c *SMSPerPhoneConfig) SetDefaults() {
	if c.Size == 0 {
		c.Size = 30
	}
	if c.ResetPeriod == "" {
		c.ResetPeriod = "24h"
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
