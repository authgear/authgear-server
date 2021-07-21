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

func (c *AuditLogFeatureConfig) SetDefaults() {
	if c.RetrievalDays == nil {
		// -1 means no limit
		c.RetrievalDays = newInt(-1)
	}
}
