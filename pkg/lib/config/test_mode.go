package config

var _ = Schema.Add("TestModeConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"oob_otp": { "$ref": "#/$defs/TestModeOOBOTPConfig" }
	}
}
`)

type TestModeConfig struct {
	FixedOOBOTP *TestModeOOBOTPConfig `json:"oob_otp,omitempty"`
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
