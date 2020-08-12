package config

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
