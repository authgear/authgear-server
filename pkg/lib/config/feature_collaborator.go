package config

var _ = FeatureConfigSchema.Add("CollaboratorFeatureConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"maximum": { "type": "integer" }
	}
}
`)

type CollaboratorFeatureConfig struct {
	Maximum *int `json:"maximum,omitempty"`
}
