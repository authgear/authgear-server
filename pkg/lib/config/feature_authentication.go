package config

var _ = FeatureConfigSchema.Add("AuthenticationFeatureConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"primary_authenticators": { "$ref": "#/$defs/AuthenticatorsFeatureConfig" },
		"secondary_authenticators": { "$ref": "#/$defs/AuthenticatorsFeatureConfig" }
	}
}
`)

type AuthenticationFeatureConfig struct {
	PrimaryAuthenticators   *AuthenticatorsFeatureConfig `json:"primary_authenticators,omitempty"`
	SecondaryAuthenticators *AuthenticatorsFeatureConfig `json:"secondary_authenticators,omitempty"`
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
