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

func NewEffectiveDefaultFeatureConfig() *FeatureConfig {
	return &FeatureConfig{}
}
