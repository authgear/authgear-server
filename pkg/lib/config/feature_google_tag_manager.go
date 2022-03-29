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
	Disabled bool `json:"disabled,omitempty"`
}
