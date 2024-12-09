package config

var _ = FeatureConfigSchema.Add("UIFeatureConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"white_labeling": { "$ref": "#/$defs/WhiteLabelingFeatureConfig" }
	}
}
`)

type UIFeatureConfig struct {
	WhiteLabeling *WhiteLabelingFeatureConfig `json:"white_labeling,omitempty"`
}

var _ MergeableFeatureConfig = &UIFeatureConfig{}

func (c *UIFeatureConfig) Merge(layer *FeatureConfig) MergeableFeatureConfig {
	if layer.UI == nil {
		return c
	}
	return layer.UI
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
