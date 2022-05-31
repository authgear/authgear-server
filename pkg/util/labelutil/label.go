package labelutil

import (
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

var minorWords = []string{
	"and",
	"but",
	"for",
	"or",
	"nor",
	"the",
	"a",
	"an",
	"to",
	"as",
}

func Label(val string) string {
	parts := strings.Split(val, "_")
	var words []string
	for i, part := range parts {
		words = append(words, titlecase(part, i, len(parts)))
	}
	return strings.Join(words, " ")
}

func titlecase(word string, index int, length int) string {
	word = strings.ToLower(word)

	shouldCapitalize := false
	if index == 0 || index == length-1 {
		shouldCapitalize = true
	} else {
		found := false
		for _, minorWord := range minorWords {
			if minorWord == word {
				found = true
			}
		}

		if found {
			shouldCapitalize = false
		} else {
			shouldCapitalize = true
		}
	}

	if shouldCapitalize {
		return cases.Title(language.English).String(word)
	}

	return word
}
