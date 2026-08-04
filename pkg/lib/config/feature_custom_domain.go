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
	// No omitempty: false is the real, correct default (not disabled), not
	// an absence -- omitempty would hide it from the effective config JSON
	// shown by the Site Admin API's feature-config UI, making a
	// fully-resolved section look empty/unset instead of explicitly
	// carrying its default value.
	Disabled bool `json:"disabled"`
}

var _ MergeableFeatureConfig = &CustomDomainFeatureConfig{}

func (c *CustomDomainFeatureConfig) Merge(layer *FeatureConfig) MergeableFeatureConfig {
	if layer.CustomDomain == nil {
		return c
	}
	return layer.CustomDomain
}
