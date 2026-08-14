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
	// No omitempty: false is this field's real default, not an absence
	// (see the update-feature-config skill).
	Disabled bool `json:"disabled"`
}

var _ MergeableFeatureConfig = &GoogleTagManagerFeatureConfig{}

func (c *GoogleTagManagerFeatureConfig) Merge(layer *FeatureConfig) MergeableFeatureConfig {
	if layer.GoogleTagManager == nil {
		return c
	}
	return layer.GoogleTagManager
}
