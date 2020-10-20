package template

import (
	"github.com/authgear/authgear-server/pkg/util/intl"
)

type LanguageItem interface {
	GetLanguageTag() string
}

func MatchLanguage(preferred []string, fallback string, items []LanguageItem) (matched LanguageItem, err error) {
	languageTagToItem := make(map[string]LanguageItem)

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

	return item, nil
}
