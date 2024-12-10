package config

var _ = FeatureConfigSchema.Add("GoogleTagManagerFeatureConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"disabled": { "type": "boolean" }
	}
}
`)

type GoogleTagManagerFeatureConfig struct {
	Disabled bool `json:"disabled,omitempty"`
}

var _ MergeableFeatureConfig = &GoogleTagManagerFeatureConfig{}

func (c *GoogleTagManagerFeatureConfig) Merge(layer *FeatureConfig) MergeableFeatureConfig {
	if layer.GoogleTagManager == nil {
		return c
	}
	return layer.GoogleTagManager
}
