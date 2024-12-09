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

var _ MergeableFeatureConfig = &AuthenticatorFeatureConfig{}

func (c *AuthenticatorFeatureConfig) Merge(layer *FeatureConfig) MergeableFeatureConfig {
	if layer.Authenticator == nil {
		return c
	}
	return layer.Authenticator
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
		"minimum_guessable_level": { "$ref": "#/$defs/PasswordPolicyItemFeatureConfig" },
		"excluded_keywords": { "$ref": "#/$defs/PasswordPolicyItemFeatureConfig" },
		"history": { "$ref": "#/$defs/PasswordPolicyItemFeatureConfig" }
	}
}
`)

type PasswordPolicyFeatureConfig struct {
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
