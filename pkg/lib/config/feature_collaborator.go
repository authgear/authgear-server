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

	merged := c
	if merged == nil {
		merged = &CollaboratorFeatureConfig{}
	}

	if layer.Collaborator.Maximum != nil {
		merged.Maximum = layer.Collaborator.Maximum
	}
	if layer.Collaborator.SoftMaximum != nil {
		merged.SoftMaximum = layer.Collaborator.SoftMaximum
	}

	return merged
}
