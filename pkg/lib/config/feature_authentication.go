package config

var _ = FeatureConfigSchema.Add("AuthenticationFeatureConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"secondary_authenticators": { "$ref": "#/$defs/AuthenticatorsFeatureConfig" }
	}
}
`)

type AuthenticationFeatureConfig struct {
	SecondaryAuthenticators *AuthenticatorsFeatureConfig `json:"secondary_authenticators,omitempty"`
}

var _ MergeableFeatureConfig = &AuthenticationFeatureConfig{}

func (c *AuthenticationFeatureConfig) Merge(layer *FeatureConfig) MergeableFeatureConfig {
	if layer.Authentication == nil {
		return c
	}

	merged := c
	if merged == nil {
		merged = &AuthenticationFeatureConfig{}
	}

	merged.SecondaryAuthenticators = merged.SecondaryAuthenticators.Merge(layer.Authentication.SecondaryAuthenticators)

	return merged
}

var _ = FeatureConfigSchema.Add("AuthenticatorsFeatureConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"oob_otp_sms": { "$ref": "#/$defs/AuthenticatorOOBOTBSMSFeatureConfig" }
	}
}
`)

type AuthenticatorsFeatureConfig struct {
	OOBOTPSMS *AuthenticatorOOBOTBSMSFeatureConfig `json:"oob_otp_sms,omitempty"`
}

func (c *AuthenticatorsFeatureConfig) Merge(layer *AuthenticatorsFeatureConfig) *AuthenticatorsFeatureConfig {
	if c == nil && layer == nil {
		return nil
	}
	if c == nil {
		return layer
	}
	if layer == nil {
		return c
	}
	if layer.OOBOTPSMS != nil {
		c.OOBOTPSMS = layer.OOBOTPSMS
	}
	return c
}

var _ = FeatureConfigSchema.Add("AuthenticatorOOBOTBSMSFeatureConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"disabled": { "type": "boolean" }
	}
}
`)

type AuthenticatorOOBOTBSMSFeatureConfig struct {
	Disabled bool `json:"disabled,omitempty"`
}
