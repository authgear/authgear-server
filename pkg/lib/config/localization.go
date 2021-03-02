package config

import "github.com/authgear/authgear-server/pkg/util/intl"

var _ = Schema.Add("LocalizationConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"fallback_language": {
			"type": "string",
			"format": "bcp47"
		},
		"supported_languages": {
			"type": "array",
			"minItems": 1,
			"uniqueItems": true,
			"items": {
				"type": "string",
				"format": "bcp47"
			}
		}
	}
}
`)

type LocalizationConfig struct {
	FallbackLanguage   *string  `json:"fallback_language,omitempty"`
	SupportedLanguages []string `json:"supported_languages,omitempty"`
}

func (c *LocalizationConfig) SetDefaults() {
	if c.FallbackLanguage == nil {
		a := intl.DefaultLanguage
		c.FallbackLanguage = &a
	}
	if c.SupportedLanguages == nil {
		c.SupportedLanguages = []string{*c.FallbackLanguage}
	}
}
