package config

// legacy code below
var _ = SecretConfigSchema.Add("Deprecated_CaptchaCloudflareCredentials", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"secret": { "type": "string" }
	},
	"required": ["secret"]
}
`)

type Deprecated_CaptchaCloudflareCredentials struct {
	Secret string `json:"secret,omitempty"`
}

func (c *Deprecated_CaptchaCloudflareCredentials) SensitiveStrings() []string {
	return []string{
		c.Secret,
	}
}
