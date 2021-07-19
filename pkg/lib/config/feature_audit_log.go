package config

var _ = FeatureConfigSchema.Add("AuditLogFeatureConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"retrieve_days": { "type": "integer" }
	}
}
`)

type AuditLogFeatureConfig struct {
	RetrieveDays *int `json:"retrieve_days,omitempty"`
}

func (c *AuditLogFeatureConfig) SetDefaults() {
	if c.RetrieveDays == nil {
		// -1 means no limit
		c.RetrieveDays = newInt(-1)
	}
}
