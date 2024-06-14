package config

var _ = SecretConfigSchema.Add("LegacyCaptchaCloudflareCredentials", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"secret": { "type": "string" }
	},
	"required": ["secret"]
}
`)

type LegacyCaptchaCloudflareCredentials struct {
	Secret string `json:"secret,omitempty"`
}

func (c *LegacyCaptchaCloudflareCredentials) SensitiveStrings() []string {
	return []string{
		c.Secret,
	}
}
