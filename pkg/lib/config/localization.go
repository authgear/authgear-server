package config

import "github.com/authgear/authgear-server/pkg/util/intl"

var _ = Schema.Add("LocalizationConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"fallback_language": { "type": "string" }
	}
}
`)

type LocalizationConfig struct {
	FallbackLanguage string `json:"fallback_language,omitempty"`
}

func (c *LocalizationConfig) SetDefaults() {
	if c.FallbackLanguage == "" {
		c.FallbackLanguage = intl.DefaultLanguage
	}
}
