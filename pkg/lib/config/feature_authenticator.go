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
	Expiry *PasswordExpiryFeatureConfig `json:"expiry,omitempty"`
}

var _ = FeatureConfigSchema.Add("PasswordPolicyFeatureConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"minimum_guessable_level": { "$ref": "#/$defs/PasswordPolicyItemFeatureConfig" },
		"excluded_keywords": { "$ref": "#/$defs/PasswordPolicyItemFeatureConfig" },
		"history": { "$ref": "#/$defs/PasswordPolicyItemFeatureConfig" },
		"expiry": { "$ref": "#/$defs/PasswordExpiryFeatureConfig" }
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

var _ = FeatureConfigSchema.Add("PasswordExpiryFeatureConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"force_change": { "$ref": "#/$defs/PasswordExpiryForceChangeFeatureConfig" }
	}
}
`)

type PasswordExpiryFeatureConfig struct {
	ForceChange *PasswordExpiryForceChangeFeatureConfig `json:"force_change,omitempty"`
}

func (c *PasswordExpiryFeatureConfig) SetDefaults() {
	if c.ForceChange == nil {
		c.ForceChange = &PasswordExpiryForceChangeFeatureConfig{}
		c.ForceChange.SetDefaults()
	}
}

var _ = FeatureConfigSchema.Add("PasswordExpiryForceChangeFeatureConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"disabled": { "type": "boolean" }
	}
}
`)

type PasswordExpiryForceChangeFeatureConfig struct {
	Disabled *bool `json:"disabled,omitempty"`
}

func (c *PasswordExpiryForceChangeFeatureConfig) SetDefaults() {
	if c.Disabled == nil {
		c.Disabled = newBool(false)
	}
}
