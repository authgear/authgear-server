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
