package config

var _ = FeatureConfigSchema.Add("AuditLogFeatureConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"retrieval_days": { "type": "integer" }
	}
}
`)

type AuditLogFeatureConfig struct {
	RetrievalDays *int `json:"retrieval_days,omitempty"`
}

var _ MergeableFeatureConfig = &AuditLogFeatureConfig{}

func (c *AuditLogFeatureConfig) Merge(layer *FeatureConfig) MergeableFeatureConfig {
	if layer.AuditLog == nil {
		return c
	}
	return layer.AuditLog
}

func (c *AuditLogFeatureConfig) SetDefaults() {
	if c.RetrievalDays == nil {
		// -1 means no limit
		c.RetrievalDays = newInt(-1)
	}
}
