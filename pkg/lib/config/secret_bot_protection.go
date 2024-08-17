package config

var _ = SecretConfigSchema.Add("BotProtectionProviderCredentials", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"type": { "type": "string", "enum": ["cloudflare", "recaptchav2"] },
		"secret_key": { "type": "string", "minLength": 1 }
	},
	"allOf": [
		{
			"if": {
				"properties": {
					"type": {
						"enum": ["cloudflare", "recaptchav2"]
					}
				},
				"required": ["type"]
			},
			"then": {
				"required": ["secret_key"]
			}
		}
	]
}
`)

type BotProtectionProviderCredentials struct {
	Type      BotProtectionProviderType `json:"type,omitempty"`
	SecretKey string                    `json:"secret_key,omitempty"`
}

func (c *BotProtectionProviderCredentials) SensitiveStrings() (sensitiveStrings []string) {
	return []string{
		c.SecretKey,
	}
}
