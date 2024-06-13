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
		"type": { "type": "string" },
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
	Type    string `json:"type,omitempty"`
	Alias   string `json:"alias,omitempty"`
	SiteKey string `json:"site_key,omitempty"` // only for some providers
}

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
