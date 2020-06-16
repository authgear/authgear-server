package config

var _ = Schema.Add("UIConfig", `
{
	"type": "object",
	"properties": {
		"custom_css": { "type": "string" },
		"country_calling_code": { "$ref": "#/$defs/UICountryCallingCodeConfig" },
		"localization": { "$ref": "#/$defs/UILocalizationConfig" }
	}
}
`)

type UIConfig struct {
	CustomCSS          string                      `json:"custom_css,omitempty"`
	CountryCallingCode *UICountryCallingCodeConfig `json:"country_calling_code,omitempty"`
	Localization       *UILocalizationConfig       `json:"localization,omitempty"`
}

var _ = Schema.Add("UICountryCallingCodeConfig", `
{
	"type": "object",
	"properties": {
		"values": { "type": "array", "items": { "type": "string" } },
		"default": { "type": "string" }
	}
}
`)

type UICountryCallingCodeConfig struct {
	Values  []string `json:"values,omitempty"`
	Default string   `json:"default,omitempty"`
}

var _ = Schema.Add("UILocalizationConfig", `
{
	"type": "object",
	"properties": {
		"fallback_language": { "type": "string" }
	}
}
`)

type UILocalizationConfig struct {
	FallbackLanguage string `json:"fallback_language,omitempty"`
}
