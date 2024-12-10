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
	Disabled bool `json:"disabled,omitempty"`
}

var _ MergeableFeatureConfig = &RateLimitsFeatureConfig{}

func (c *RateLimitsFeatureConfig) Merge(layer *FeatureConfig) MergeableFeatureConfig {
	if layer.RateLimits == nil {
		return c
	}
	return layer.RateLimits
}
