package config

var _ = Schema.Add("CaptchaConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"enabled": { "type": "boolean" },
		"providers": {
			"type": "array",
			"items": { "$ref": "#/$defs/CaptchaProvider" }
		},
		"provider": { "$ref": "#/$defs/Deprecated_CaptchaProvider" }
	}
}
`)

var _ = Schema.Add("CaptchaProvider", `
{
	"type": "object",
	"additionalProperties": false,
	"required": ["type", "alias"],
	"properties": {
		"type": { "type": "string", "enum": ["cloudflare", "recaptchav2"] },
		"alias": { "type": "string" },
		"site_key": { "type": "string" }
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
				"required": ["site_key"]
			}
		}
	]
}
`)

type CaptchaConfig struct {
	Deprecated_Provider *Deprecated_CaptchaProvider `json:"provider,omitempty"`
	Enabled             bool                        `json:"enabled,omitempty"`
	Providers           []*CaptchaProvider          `json:"providers,omitempty"`
}

type CaptchaProvider struct {
	Type    CaptchaProviderType `json:"type,omitempty"`
	Alias   string              `json:"alias,omitempty"`
	SiteKey string              `json:"site_key,omitempty"` // only for some providers
}

type CaptchaProviderType string

const (
	CaptchaProviderTypeCloudflare  CaptchaProviderType = "cloudflare"
	CaptchaProviderTypeRecaptchaV2 CaptchaProviderType = "recaptchav2"
)

// legacy code below

var _ = Schema.Add("Deprecated_CaptchaProvider", `
{
	"type": "string",
	"enum": ["cloudflare"]
}
`)

type Deprecated_CaptchaProvider string

const (
	Deprecated_CaptchaProviderCloudflare Deprecated_CaptchaProvider = "cloudflare"
)
