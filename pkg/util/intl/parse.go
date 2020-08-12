package intl

import (
	"strings"

	"golang.org/x/text/language"
)

func ParseUILocales(uiLocales string) []string {
	// Convert ui_locales into Accept-Language header format
	acceptLanguage := strings.ReplaceAll(uiLocales, " ", ", ")

	return ParseAcceptLanguage(acceptLanguage)
}

func ParseAcceptLanguage(header string) []string {
	tags, _, err := language.ParseAcceptLanguage(header)
	if err != nil {
		return nil
	}

	var out []string
	for _, tag := range tags {
		out = append(out, tag.String())
	}
	return out
}
