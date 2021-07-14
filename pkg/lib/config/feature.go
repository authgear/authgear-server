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
		"identity": { "$ref": "#/$defs/IdentityFeatureConfig" },
		"authentication": { "$ref": "#/$defs/AuthenticationFeatureConfig" },
		"custom_domain": { "$ref": "#/$defs/CustomDomainFeatureConfig" },
		"ui": { "$ref": "#/$defs/UIFeatureConfig" },
		"oauth": { "$ref": "#/$defs/OAuthFeatureConfig" }
	}
}
`)

type FeatureConfig struct {
	Identity       *IdentityFeatureConfig       `json:"identity,omitempty"`
	Authentication *AuthenticationFeatureConfig `json:"authentication,omitempty"`
	CustomDomain   *CustomDomainFeatureConfig   `json:"custom_domain,omitempty"`
	UI             *UIFeatureConfig             `json:"ui,omitempty"`
	OAuth          *OAuthFeatureConfig          `json:"oauth,omitempty"`
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

	setFieldDefaults(&config)

	return &config, nil
}

func NewEffectiveDefaultFeatureConfig() *FeatureConfig {
	config := FeatureConfig{}
	setFieldDefaults(&config)
	return &config
}
