package config

var _ = FeatureConfigSchema.Add("AuthenticatorFeatureConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"password": { "$ref": "#/$defs/AuthenticatorPasswordFeatureConfig" }
	}
}
`)

type AuthenticatorFeatureConfig struct {
	Password *AuthenticatorPasswordFeatureConfig `json:"password,omitempty"`
}

var _ = FeatureConfigSchema.Add("AuthenticatorPasswordFeatureConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"policy": { "$ref": "#/$defs/PasswordPolicyFeatureConfig" }
	}
}
`)

type AuthenticatorPasswordFeatureConfig struct {
	Policy *PasswordPolicyFeatureConfig `json:"policy,omitempty"`
}

var _ = FeatureConfigSchema.Add("PasswordPolicyFeatureConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"min_length": { "$ref": "#/$defs/PasswordPolicyItemFeatureConfig" },
		"uppercase_required": { "$ref": "#/$defs/PasswordPolicyItemFeatureConfig" },
		"lowercase_required": { "$ref": "#/$defs/PasswordPolicyItemFeatureConfig" },
		"digit_required": { "$ref": "#/$defs/PasswordPolicyItemFeatureConfig" },
		"symbol_required": { "$ref": "#/$defs/PasswordPolicyItemFeatureConfig" },
		"minimum_guessable_level": { "$ref": "#/$defs/PasswordPolicyItemFeatureConfig" },
		"excluded_keywords": { "$ref": "#/$defs/PasswordPolicyItemFeatureConfig" },
		"history": { "$ref": "#/$defs/PasswordPolicyItemFeatureConfig" }
	}
}
`)

type PasswordPolicyFeatureConfig struct {
	MinLength             *PasswordPolicyItemFeatureConfig `json:"min_length,omitempty"`
	UppercaseRequired     *PasswordPolicyItemFeatureConfig `json:"uppercase_required,omitempty"`
	LowercaseRequired     *PasswordPolicyItemFeatureConfig `json:"lowercase_required,omitempty"`
	DigitRequired         *PasswordPolicyItemFeatureConfig `json:"digit_required,omitempty"`
	SymbolRequired        *PasswordPolicyItemFeatureConfig `json:"symbol_required,omitempty"`
	MinimumGuessableLevel *PasswordPolicyItemFeatureConfig `json:"minimum_guessable_level,omitempty"`
	ExcludedKeywords      *PasswordPolicyItemFeatureConfig `json:"excluded_keywords,omitempty"`
	History               *PasswordPolicyItemFeatureConfig `json:"history,omitempty"`
}

var _ = FeatureConfigSchema.Add("PasswordPolicyItemFeatureConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"disabled": { "type": "boolean" }
	}
}
`)

type PasswordPolicyItemFeatureConfig struct {
	Disabled *bool `json:"disabled,omitempty"`
}

func (c *PasswordPolicyItemFeatureConfig) SetDefaults() {
	if c.Disabled == nil {
		c.Disabled = newBool(false)
	}
}
