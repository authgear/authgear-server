package config

var _ = Schema.Add("CaptchaConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"provider": { "$ref": "#/$defs/CaptchaProvider" }
	},
	"required": ["provider"]
}
`)

type CaptchaConfig struct {
	Provider *CaptchaProvider `json:"provider,omitempty"`
}

var _ = Schema.Add("CaptchaProvider", `
{
	"type": "string",
	"enum": ["cloudflare"]
}
`)

type CaptchaProvider string

const (
	CaptchaProviderCloudflare CaptchaProvider = "cloudflare"
)
