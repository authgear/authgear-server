package config

var _ = Schema.Add("CaptchaConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"provider": { "$ref": "#/$defs/LegacyCaptchaProvider" }
	}
}
`)

type CaptchaConfig struct {
	LegacyProvider *LegacyCaptchaProvider `json:"provider,omitempty"`
}

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
