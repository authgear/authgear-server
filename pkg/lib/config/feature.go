package config

import (
	"bytes"
	"encoding/json"

	"sigs.k8s.io/yaml"
)

var _ = FeatureConfigSchema.Add("FeatureConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"authentication": { "$ref": "#/$defs/AuthenticationFeatureConfig" }
	}
}
`)

type FeatureConfig struct {
	Authentication *AuthenticationFeatureConfig `json:"authentication,omitempty"`
}

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

func ParseFeatureConfig(inputYAML []byte) (*FeatureConfig, error) {
	jsonData, err := yaml.YAMLToJSON(inputYAML)
	if err != nil {
		return nil, err
	}

	err = FeatureConfigSchema.Validator().ValidateWithMessage(
		bytes.NewReader(jsonData),
		"invalid feature config",
	)
	if err != nil {
		return nil, err
	}

	var config FeatureConfig
	decoder := json.NewDecoder(bytes.NewReader(jsonData))
	err = decoder.Decode(&config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}
