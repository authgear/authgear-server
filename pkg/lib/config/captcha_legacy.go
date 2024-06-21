package config

var _ = Schema.Add("CaptchaConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"provider": { "$ref": "#/$defs/Deprecated_CaptchaProvider" }
	}
}
`)

type CaptchaConfig struct {
	Deprecated_Provider *Deprecated_CaptchaProvider `json:"provider,omitempty"`
}

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
