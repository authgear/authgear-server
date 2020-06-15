package config

type UIConfig struct {
	CustomCSS          string                      `json:"custom_css,omitempty"`
	CountryCallingCode *UICountryCallingCodeConfig `json:"country_calling_code,omitempty"`
	Localization       *UILocalizationConfig       `json:"localization,omitempty"`
}

type UICountryCallingCodeConfig struct {
	Values  []string `json:"values,omitempty"`
	Default string   `json:"default,omitempty"`
}

type UILocalizationConfig struct {
	FallbackLanguage string `json:"fallback_language,omitempty"`
}
