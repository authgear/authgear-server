package config

var _ = FeatureConfigSchema.Add("TestModeFeatureConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"fixed_oob_otp": { "$ref": "#/$defs/TestModeFixedOOBOTPConfig" },
		"deterministic_link_otp": { "$ref": "#/$defs/TestModeDeterministicLinkOTPConfig" }
	}
}
`)

var _ = FeatureConfigSchema.Add("TestModeFixedOOBOTPConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"enabled": { "type": "boolean" },
		"code": { "type": "string" }
	},
	"required": ["enabled", "code"]
}
`)

var _ = FeatureConfigSchema.Add("TestModeDeterministicLinkOTPConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"enabled": { "type": "boolean" }
	},
	"required": ["enabled"]
}
`)

type TestModeFeatureConfig struct {
	FixedOOBOTP          *TestModeFixedOOBOTPFeatureConfig          `json:"fixed_oob_otp,omitempty"`
	DeterministicLinkOTP *TestModeDeterministicLinkOTPFeatureConfig `json:"deterministic_link_otp,omitempty"`
}

type TestModeFixedOOBOTPFeatureConfig struct {
	Enabled bool   `json:"enabled"`
	Code    string `json:"code"`
}

type TestModeDeterministicLinkOTPFeatureConfig struct {
	Enabled bool `json:"enabled"`
}
