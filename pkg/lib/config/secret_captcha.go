package config

var _ = SecretConfigSchema.Add("CaptchaCloudflareCredentials", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"secret": { "type": "string" }
	},
	"required": ["secret"]
}
`)

type CaptchaCloudflareCredentials struct {
	Secret string `json:"secret,omitempty"`
}

func (c *CaptchaCloudflareCredentials) SensitiveStrings() []string {
	return []string{
		c.Secret,
	}
}
