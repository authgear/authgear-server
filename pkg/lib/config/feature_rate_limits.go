package config

var _ = FeatureConfigSchema.Add("RateLimitsFeatureConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"disabled": { "type": "boolean" }
	}
}
`)

type RateLimitsFeatureConfig struct {
	// No omitempty: false is this field's real default, not an absence
	// (see the update-feature-config skill).
	Disabled bool `json:"disabled"`
}

var _ MergeableFeatureConfig = &RateLimitsFeatureConfig{}

func (c *RateLimitsFeatureConfig) Merge(layer *FeatureConfig) MergeableFeatureConfig {
	if layer.RateLimits == nil {
		return c
	}
	return layer.RateLimits
}
