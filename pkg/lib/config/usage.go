package config

import "github.com/authgear/authgear-server/pkg/api/model"

var _ = Schema.Add("UsageLimitPeriod", `
{
	"type": "string",
	"enum": ["day", "month"]
}
`)

var _ = Schema.Add("UsageLimitAction", `
{
	"type": "string",
	"enum": ["alert", "block"]
}
`)

var _ = Schema.Add("UsageMatch", `
{
	"type": "string",
	"enum": ["*", "user_export", "user_import", "email", "whatsapp", "sms"]
}
`)

var _ = Schema.Add("UsageLimitConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"quota": { "type": "integer", "minimum": 0 },
		"period": { "$ref": "#/$defs/UsageLimitPeriod" },
		"action": { "$ref": "#/$defs/UsageLimitAction" }
	},
	"required": ["quota", "period", "action"]
}
`)

type UsageLimitConfig struct {
	Quota  int                    `json:"quota"`
	Period model.UsageLimitPeriod `json:"period"`
	Action model.UsageLimitAction `json:"action"`
}

var _ = Schema.Add("UsageLimitsConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"user_export": { "type": "array", "items": { "$ref": "#/$defs/UsageLimitConfig" } },
		"user_import": { "type": "array", "items": { "$ref": "#/$defs/UsageLimitConfig" } },
		"email": { "type": "array", "items": { "$ref": "#/$defs/UsageLimitConfig" } },
		"whatsapp": { "type": "array", "items": { "$ref": "#/$defs/UsageLimitConfig" } },
		"sms": { "type": "array", "items": { "$ref": "#/$defs/UsageLimitConfig" } }
	}
}
`)

type UsageLimitsConfig struct {
	UserExport []UsageLimitConfig `json:"user_export,omitempty"`
	UserImport []UsageLimitConfig `json:"user_import,omitempty"`
	Email      []UsageLimitConfig `json:"email,omitempty"`
	Whatsapp   []UsageLimitConfig `json:"whatsapp,omitempty"`
	SMS        []UsageLimitConfig `json:"sms,omitempty"`
}

func (c *UsageLimitsConfig) Limits(name model.UsageName) []UsageLimitConfig {
	if c == nil {
		return nil
	}

	switch name {
	case model.UsageNameUserExport:
		return c.UserExport
	case model.UsageNameUserImport:
		return c.UserImport
	case model.UsageNameEmail:
		return c.Email
	case model.UsageNameWhatsapp:
		return c.Whatsapp
	case model.UsageNameSMS:
		return c.SMS
	default:
		return nil
	}
}

var _ = Schema.Add("UsageAlertConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"type": { "type": "string", "enum": ["email"] },
		"email": { "type": "string", "format": "email" },
		"match": { "$ref": "#/$defs/UsageMatch" }
	},
	"required": ["type", "match"],
	"allOf": [
		{
			"if": {
				"properties": {
					"type": { "const": "email" }
				},
				"required": ["type"]
			},
			"then": {
				"required": ["email"]
			}
		}
	]
}
`)

type UsageAlertConfig struct {
	Type  string `json:"type"`
	Email string `json:"email,omitempty"`
	Match string `json:"match"`
}

var _ = Schema.Add("UsageConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"alerts": { "type": "array", "items": { "$ref": "#/$defs/UsageAlertConfig" } },
		"limits": { "$ref": "#/$defs/UsageLimitsConfig" }
	}
}
`)

type UsageConfig struct {
	Alerts []UsageAlertConfig `json:"alerts,omitempty"`
	Limits *UsageLimitsConfig `json:"limits,omitempty" nullable:"true"`
}
