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
		"provider": { "$ref": "#/$defs/LegacyCaptchaProvider" }
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
	}
}
`)

type CaptchaConfig struct {
	LegacyProvider *LegacyCaptchaProvider `json:"provider,omitempty"`
	Enabled        bool                   `json:"enabled,omitempty"`
	Providers      []*CaptchaProvider     `json:"providers,omitempty"`
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

var _ = Schema.Add("LegacyCaptchaProvider", `
{
	"type": "string",
	"enum": ["cloudflare"]
}
`)

type LegacyCaptchaProvider string

const (
	LegacyCaptchaProviderCloudflare LegacyCaptchaProvider = "cloudflare"
)
