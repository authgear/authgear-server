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

	merged := c
	if merged == nil {
		merged = &AuthenticatorFeatureConfig{}
	}

	merged.Password = merged.Password.Merge(layer.Authenticator.Password)

	return merged
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

func (c *AuthenticatorPasswordFeatureConfig) Merge(layer *AuthenticatorPasswordFeatureConfig) *AuthenticatorPasswordFeatureConfig {
	if c == nil && layer == nil {
		return nil
	}
	if c == nil {
		return layer
	}
	if layer == nil {
		return c
	}
	c.Policy = c.Policy.Merge(layer.Policy)
	return c
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

func (c *PasswordPolicyFeatureConfig) Merge(layer *PasswordPolicyFeatureConfig) *PasswordPolicyFeatureConfig {
	if c == nil && layer == nil {
		return nil
	}
	if c == nil {
		return layer
	}
	if layer == nil {
		return c
	}
	if layer.MinimumGuessableLevel != nil {
		c.MinimumGuessableLevel = layer.MinimumGuessableLevel
	}
	if layer.ExcludedKeywords != nil {
		c.ExcludedKeywords = layer.ExcludedKeywords
	}
	if layer.History != nil {
		c.History = layer.History
	}
	return c
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
		c.Disabled = new(false)
	}
}
