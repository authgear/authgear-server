package config

var _ = FeatureConfigSchema.Add("CollaboratorFeatureConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"maximum": { "type": "integer" },
		"soft_maximum": { "type": "integer" }
	}
}
`)

type CollaboratorFeatureConfig struct {
	Maximum     *int `json:"maximum,omitempty"`
	SoftMaximum *int `json:"soft_maximum,omitempty"`
}

var _ MergeableFeatureConfig = &CollaboratorFeatureConfig{}

func (c *CollaboratorFeatureConfig) Merge(layer *FeatureConfig) MergeableFeatureConfig {
	if layer.Collaborator == nil {
		return c
	}
	return layer.Collaborator
}
