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

// SupportedLanguages ensures fallback language is the first element.
type SupportedLanguages []string

// Supported constructs SupportedLanguages.
func Supported(supportedLanguageTags []string, fallbackLanguage FallbackLanguage) SupportedLanguages {
	fallbackLanguageTag := string(fallbackLanguage)

	fallbackIdx := -1
	for i, tag := range supportedLanguageTags {
		if tag == fallbackLanguageTag {
			fallbackIdx = i
			break
		}
	}

	if fallbackIdx < 0 {
		var s []string
		s = append(s, fallbackLanguageTag)
		s = append(s, supportedLanguageTags[:]...)
		return s
	}

	var s []string
	s = append(s, fallbackLanguageTag)
	s = append(s, supportedLanguageTags[:fallbackIdx]...)
	s = append(s, supportedLanguageTags[fallbackIdx+1:]...)
	return s
}
