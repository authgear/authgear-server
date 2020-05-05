package intl

// DefaultLanguage is the default language. It is english.
// Default templates and translation are written in english.
const DefaultLanguage = "en"

// FallbackLanguage is non-empty.
// Its purpose is to ensure fallback language is non-empty at compile time.
// Use Fallback to construct FallbackLanguage.
// Use string() to convert it back to a string.
type FallbackLanguage string

// Fallback constructs FallbackLanguage.
func Fallback(fallbackLanguageTag string) FallbackLanguage {
	if fallbackLanguageTag == "" {
		return FallbackLanguage(DefaultLanguage)
	}
	return FallbackLanguage(fallbackLanguageTag)
}
