package config

import "github.com/authgear/authgear-server/pkg/util/phone"

var _ = FeatureConfigSchema.Add("UIFeatureConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"white_labeling": { "$ref": "#/$defs/WhiteLabelingFeatureConfig" },
		"phone_input": { "$ref": "#/$defs/PhoneInputFeatureConfig" }
	}
}
`)

type UIFeatureConfig struct {
	WhiteLabeling *WhiteLabelingFeatureConfig `json:"white_labeling,omitempty"`
	PhoneInput    *PhoneInputFeatureConfig    `json:"phone_input,omitempty"`
}

var _ MergeableFeatureConfig = &UIFeatureConfig{}

func (c *UIFeatureConfig) Merge(layer *FeatureConfig) MergeableFeatureConfig {
	if layer.UI == nil {
		return c
	}

	merged := c
	if merged == nil {
		merged = &UIFeatureConfig{}
	}

	if layer.UI.WhiteLabeling != nil {
		merged.WhiteLabeling = layer.UI.WhiteLabeling
	}
	if layer.UI.PhoneInput != nil {
		merged.PhoneInput = layer.UI.PhoneInput
	}

	return merged
}

var _ = FeatureConfigSchema.Add("WhiteLabelingFeatureConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"disabled": { "type": "boolean" }
	}
}
`)

type WhiteLabelingFeatureConfig struct {
	Disabled bool `json:"disabled,omitempty"`
}

var _ = FeatureConfigSchema.Add("PhoneInputFeatureConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"allowlist": { "type": "array", "items": { "$ref": "#/$defs/ISO31661Alpha2" } }
	}
}
`)

type PhoneInputFeatureConfig struct {
	AllowList []string `json:"allowlist,omitempty"`
}

var _ = FeatureConfigSchema.Add("ISO31661Alpha2", phone.JSONSchemaString)
