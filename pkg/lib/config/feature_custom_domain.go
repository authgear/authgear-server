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
	Disabled bool `json:"disabled,omitempty"`
}

var _ MergeableFeatureConfig = &CustomDomainFeatureConfig{}

func (c *CustomDomainFeatureConfig) Merge(layer *FeatureConfig) MergeableFeatureConfig {
	if layer.CustomDomain == nil {
		return c
	}
	return layer.CustomDomain
}
