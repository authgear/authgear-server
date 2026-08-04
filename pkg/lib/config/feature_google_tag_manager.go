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
	// No omitempty: see the comment on CustomDomainFeatureConfig.Disabled.
	Disabled bool `json:"disabled"`
}

var _ MergeableFeatureConfig = &GoogleTagManagerFeatureConfig{}

func (c *GoogleTagManagerFeatureConfig) Merge(layer *FeatureConfig) MergeableFeatureConfig {
	if layer.GoogleTagManager == nil {
		return c
	}
	return layer.GoogleTagManager
}
