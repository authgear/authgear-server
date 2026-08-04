package config

var _ = FeatureConfigSchema.Add("CustomDomainFeatureConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"disabled": { "type": "boolean" }
	}
}
`)

type CustomDomainFeatureConfig struct {
	// No omitempty: false is this field's real default, not an absence
	// (see the update-feature-config skill).
	Disabled bool `json:"disabled"`
}

var _ MergeableFeatureConfig = &CustomDomainFeatureConfig{}

func (c *CustomDomainFeatureConfig) Merge(layer *FeatureConfig) MergeableFeatureConfig {
	if layer.CustomDomain == nil {
		return c
	}
	return layer.CustomDomain
}
