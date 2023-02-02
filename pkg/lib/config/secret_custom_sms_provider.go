package config

var _ = SecretConfigSchema.Add("CustomSMSProviderConfigs", `
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

type CustomSMSProviderConfigs struct {
	URL     string           `json:"url,omitempty"`
	Timeout *DurationSeconds `json:"timeout,omitempty"`
}

func (c *CustomSMSProviderConfigs) SensitiveStrings() []string {
	return []string{
		c.URL,
	}
}
