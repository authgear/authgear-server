package config

var _ = SecretConfigSchema.Add("CustomSMSProviderConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"url": { "type": "string" },
		"timeout": { "type": "integer" }
	},
	"required": ["url"]
}
`)

type CustomSMSProviderConfig struct {
	URL     string           `json:"url,omitempty"`
	Timeout *DurationSeconds `json:"timeout,omitempty"`
}

func (c *CustomSMSProviderConfig) SensitiveStrings() []string {
	return []string{
		c.URL,
	}
}
