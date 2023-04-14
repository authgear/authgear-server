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
