package intl

import (
	"golang.org/x/text/language"
)

// Resolve resolved language based on fallback and supportedLanguages config.
// Return index of supportedLanguages and resolved language tag.
// Return -1 if not found
func Resolve(preferred []string, fallback string, supported []string) (int, language.Tag) {
	supportedLanguageTags := Supported(supported, Fallback(fallback))
	supportedLanguagesIdx := map[string]int{}
	for i, item := range supported {
		supportedLanguagesIdx[item] = i
	}

	idx, tag := Match(preferred, supportedLanguageTags)
	if idx == -1 {
		return idx, tag
	}

	matched := supportedLanguageTags[idx]
	if idx, ok := supportedLanguagesIdx[matched]; ok {
		return idx, tag
	}

	return -1, tag
}

func ResolveLocaleCode(resolved string, fallback string, supported []string) string {
	var matcher = language.NewMatcher(SupportedLanguageTags(supported))
	var locale language.Tag

	locale, _, confidence := matcher.Match(language.MustParse(resolved))
	if confidence == language.No {
		locale, _ = language.Parse(fallback)
	}

	_, _, region := locale.Raw()
	localeCode := locale.String()
	if locale.Parent() != locale {
		localeCode = locale.Parent().String() + "-" + region.String()
	}

	return localeCode
}
