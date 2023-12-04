package config

import "regexp"

var _ = Schema.Add("TestModeConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"oob_otp": { "$ref": "#/$defs/TestModeOOBOTPConfig" },
		"sms": { "$ref": "#/$defs/TestModeSMSConfig" },
		"whatsapp": { "$ref": "#/$defs/TestModeWhatsappConfig" },
		"email": { "$ref": "#/$defs/TestModeEmailConfig" }
	}
}
`)

type TestModeConfig struct {
	FixedOOBOTP *TestModeOOBOTPConfig   `json:"oob_otp,omitempty"`
	SMS         *TestModeSMSConfig      `json:"sms,omitempty"`
	Whatsapp    *TestModeWhatsappConfig `json:"whatsapp,omitempty"`
	Email       *TestModeEmailConfig    `json:"email,omitempty"`
}

var _ = Schema.Add("TestModeOOBOTPConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"enabled": { "type": "boolean" },
		"rules": { "type": "array", "items": { "$ref": "#/$defs/TestModeOOBOTPRule" } }
	}
}
`)

type TestModeOOBOTPConfig struct {
	Enabled bool                  `json:"enabled,omitempty"`
	Rules   []*TestModeOOBOTPRule `json:"rules,omitempty"`
}

var _ = Schema.Add("TestModeOOBOTPRule", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"regex": { "type": "string", "format": "x_re2_regex" },
		"fixed_code": { "type": "string" }
	},
	"required": ["regex"]
}
`)

type TestModeOOBOTPRule struct {
	Regex     string `json:"regex,omitempty"`
	FixedCode string `json:"fixed_code,omitempty"`
}

func (r *TestModeOOBOTPRule) GetRegex() *regexp.Regexp {
	return regexp.MustCompile(r.Regex)
}

var _ = Schema.Add("TestModeSMSConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"enabled": { "type": "boolean" },
		"rules": { "type": "array", "items": { "$ref": "#/$defs/TestModeSMSRule" } }
	}
}
`)

type TestModeSMSConfig struct {
	Enabled bool               `json:"enabled,omitempty"`
	Rules   []*TestModeSMSRule `json:"rules,omitempty"`
}

var _ = Schema.Add("TestModeSMSRule", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"regex": { "type": "string", "format": "x_re2_regex" },
		"suppressed": { "type": "boolean" }
	},
	"required": ["regex"]
}
`)

type TestModeSMSRule struct {
	Regex      string `json:"regex,omitempty"`
	Suppressed string `json:"suppressed,omitempty"`
}

func (r *TestModeSMSRule) GetRegex() *regexp.Regexp {
	return regexp.MustCompile(r.Regex)
}

var _ = Schema.Add("TestModeWhatsappConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"enabled": { "type": "boolean" },
		"rules": { "type": "array", "items": { "$ref": "#/$defs/TestModeWhatsappRule" } }
	}
}
`)

type TestModeWhatsappConfig struct {
	Enabled bool                    `json:"enabled,omitempty"`
	Rules   []*TestModeWhatsappRule `json:"rules,omitempty"`
}

var _ = Schema.Add("TestModeWhatsappRule", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"regex": { "type": "string", "format": "x_re2_regex" },
		"suppressed": { "type": "boolean" }
	},
	"required": ["regex"]
}
`)

type TestModeWhatsappRule struct {
	Regex      string `json:"regex,omitempty"`
	Suppressed string `json:"suppressed,omitempty"`
}

func (r *TestModeWhatsappRule) GetRegex() *regexp.Regexp {
	return regexp.MustCompile(r.Regex)
}

var _ = Schema.Add("TestModeEmailConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"enabled": { "type": "boolean" },
		"rules": { "type": "array", "items": { "$ref": "#/$defs/TestModeEmailRule" } }
	}
}
`)

type TestModeEmailConfig struct {
	Enabled bool                 `json:"enabled,omitempty"`
	Rules   []*TestModeEmailRule `json:"rules,omitempty"`
}

var _ = Schema.Add("TestModeEmailRule", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"regex": { "type": "string", "format": "x_re2_regex" },
		"suppressed": { "type": "boolean" }
	},
	"required": ["regex"]
}
`)

type TestModeEmailRule struct {
	Regex      string `json:"regex,omitempty"`
	Suppressed string `json:"suppressed,omitempty"`
}

func (r *TestModeEmailRule) GetRegex() *regexp.Regexp {
	return regexp.MustCompile(r.Regex)
}
