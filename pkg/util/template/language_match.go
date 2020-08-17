package template

import (
	"github.com/authgear/authgear-server/pkg/util/intl"
)

type languageTagger interface {
	GetLanguageTag() string
}

func languageMatch(preferred []string, fallback string, items []languageTagger) (matched *languageTagger, err error) {
	languageTagToItem := make(map[string]languageTagger)

	var rawSupported []string
	for _, item := range items {
		tag := item.GetLanguageTag()
		languageTagToItem[tag] = item
		rawSupported = append(rawSupported, tag)
	}
	supportedLanguageTags := intl.Supported(rawSupported, intl.Fallback(fallback))

	idx, _ := intl.Match(preferred, supportedLanguageTags)
	tag := supportedLanguageTags[idx]

	item, ok := languageTagToItem[tag]
	if !ok {
		err = ErrNoLanguageMatch
		return
	}

	return &item, nil
}
