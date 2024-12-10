package config

var _ = FeatureConfigSchema.Add("TestModeFeatureConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"fixed_oob_otp": { "$ref": "#/$defs/TestModeFixedOOBOTPConfig" },
		"deterministic_link_otp": { "$ref": "#/$defs/TestModeDeterministicLinkOTPConfig" },
		"sms": { "$ref": "#/$defs/TestModeSMSFeatureConfig" },
		"whatsapp": { "$ref": "#/$defs/TestModeWhatsappFeatureConfig" },
		"email": { "$ref": "#/$defs/TestModeEmailFeatureConfig" }
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

var _ = FeatureConfigSchema.Add("TestModeSMSFeatureConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"suppressed": { "type": "boolean" }
	}
}
`)

var _ = FeatureConfigSchema.Add("TestModeWhatsappFeatureConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"suppressed": { "type": "boolean" }
	}
}
`)

var _ = FeatureConfigSchema.Add("TestModeEmailFeatureConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"suppressed": { "type": "boolean" }
	}
}
`)

type TestModeFeatureConfig struct {
	FixedOOBOTP          *TestModeFixedOOBOTPFeatureConfig          `json:"fixed_oob_otp,omitempty"`
	DeterministicLinkOTP *TestModeDeterministicLinkOTPFeatureConfig `json:"deterministic_link_otp,omitempty"`
	SMS                  *TestModeSMSFeatureConfig                  `json:"sms,omitempty"`
	Whatsapp             *TestModeWhatsappFeatureConfig             `json:"whatsapp,omitempty"`
	Email                *TestModeEmailFeatureConfig                `json:"email,omitempty"`
}

var _ MergeableFeatureConfig = &TestModeFeatureConfig{}

func (c *TestModeFeatureConfig) Merge(layer *FeatureConfig) MergeableFeatureConfig {
	if layer.TestMode == nil {
		return c
	}
	return layer.TestMode
}

type TestModeFixedOOBOTPFeatureConfig struct {
	Enabled bool   `json:"enabled"`
	Code    string `json:"code"`
}

type TestModeDeterministicLinkOTPFeatureConfig struct {
	Enabled bool `json:"enabled"`
}

type TestModeSMSFeatureConfig struct {
	Suppressed bool `json:"suppressed"`
}

type TestModeEmailFeatureConfig struct {
	Suppressed bool `json:"suppressed"`
}

type TestModeWhatsappFeatureConfig struct {
	Suppressed bool `json:"suppressed"`
}

type FeatureTestModeSMSSuppressed bool

type FeatureTestModeWhatsappSuppressed bool

type FeatureTestModeEmailSuppressed bool
