package config

//go:generate msgp -tests=false

type LocalizationConfiguration struct {
	FallbackLanguage string `json:"fallback_language,omitempty" yaml:"fallback_language" msg:"fallback_language"`
}
