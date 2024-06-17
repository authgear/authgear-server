package config

var _ = SecretConfigSchema.Add("CaptchaProvidersCredentials", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"items": {
			"type": "array",
			"items": { "$ref": "#/$defs/CaptchaProvidersCredentialsItem" }
		}
	}
}
`)

var _ = SecretConfigSchema.Add("CaptchaProvidersCredentialsItem", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"type": { "type": "string", "enum": ["cloudflare", "recaptchav2"] },
		"alias": { "type": "string" },
		"secret_key": { "type": "string" }
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

type CaptchaProvidersCredentials struct {
	Items []CaptchaProvidersCredentialsItem `json:"items,omitempty"`
}

type CaptchaProvidersCredentialsItem struct {
	Type      CaptchaProvidersCredentialsItemType `json:"type,omitempty"`
	Alias     string                              `json:"alias,omitempty"`
	SecretKey string                              `json:"secret_key,omitempty"`
}

type CaptchaProvidersCredentialsItemType string

const (
	CaptchaProvidersCredentialsItemTypeCloudflare  CaptchaProvidersCredentialsItemType = "cloudflare"
	CaptchaProvidersCredentialsItemTypeRecaptchaV2 CaptchaProvidersCredentialsItemType = "recaptchav2"
)

func (c *CaptchaProvidersCredentials) SensitiveStrings() []string {
	sensitiveStrings := make([]string, len(c.Items))
	for i, cred := range c.Items {
		sensitiveStrings[i] = cred.SecretKey
	}
	return sensitiveStrings
}

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
