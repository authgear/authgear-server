package intl

import (
	"golang.org/x/text/language"
)

// Match matches preferredLanguageTags to supportedLanguageTags
// using fallbackLanguageTag as fallback.
func Match(preferredLanguageTags []string, supportedLanguageTags SupportedLanguages) (int, language.Tag) {
	if len(supportedLanguageTags) <= 0 {
		return -1, language.Und
	}

	supported := make([]language.Tag, len(supportedLanguageTags))
	for i, item := range supportedLanguageTags {
		supported[i] = language.Make(item)
	}
	matcher := language.NewMatcher(supported)

	preferred := make([]language.Tag, len(preferredLanguageTags))
	for i, tagStr := range preferredLanguageTags {
		preferred[i] = language.Make(tagStr)
	}

	_, idx, _ := matcher.Match(preferred...)
	tag := supported[idx]

	return idx, tag
}
