package pgsearch

import (
	"strings"

	"github.com/rivo/uniseg"
)

func MapUnicodeSegmentation(original map[string]string) map[string]string {
	new := map[string]string{}
	for k, v := range original {
		new[k] = StringUnicodeSegmentation(v)
	}
	return new
}

func StringUnicodeSegmentation(original string) string {
	state := -1
	str := original
	words := []string{}
	for len(str) > 0 {
		var word string
		word, str, state = uniseg.FirstWordInString(str, state)
		words = append(words, word)
	}
	return strings.Join(words, " ")
}
