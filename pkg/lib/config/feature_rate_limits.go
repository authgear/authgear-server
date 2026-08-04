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
	// No omitempty: see the comment on CustomDomainFeatureConfig.Disabled.
	Disabled bool `json:"disabled"`
}

var _ MergeableFeatureConfig = &RateLimitsFeatureConfig{}

func (c *RateLimitsFeatureConfig) Merge(layer *FeatureConfig) MergeableFeatureConfig {
	if layer.RateLimits == nil {
		return c
	}
	return layer.RateLimits
}
