package nameutil

import (
	"strings"

	"github.com/abadojack/whatlanggo"
)

func isCJK(lang whatlanggo.Lang) bool {
	switch lang {
	case whatlanggo.Cmn: // Chinese
		fallthrough
	case whatlanggo.Jpn: // Japanese
		fallthrough
	case whatlanggo.Kor: // Korean
		return true
	default:
		return false
	}
}

func isEasternNameOrder(lang whatlanggo.Lang) bool {
	// https://en.wikipedia.org/wiki/Personal_name#Eastern_name_order
	switch lang {
	case whatlanggo.Cmn: // Chinese
		fallthrough
	case whatlanggo.Jpn: // Japanese
		fallthrough
	case whatlanggo.Kor: // Korean
		fallthrough
	case whatlanggo.Vie: // Vietnamese
		fallthrough
	case whatlanggo.Khm: // Khmer (Cambodia)
		return true
	default:
		return false
	}
}

// Format formats givenName, middleName, familyName with best effort.
// One caveat is that it cannot format name in Hong Kong style.
func Format(givenName, middleName, familyName string) (out string) {
	forTest := strings.Join([]string{givenName, middleName, familyName}, " ")
	info := whatlanggo.Detect(forTest)

	// Assume English.
	lang := whatlanggo.Eng
	if info.IsReliable() {
		lang = info.Lang
	}

	westernOrdered := []string{givenName, middleName, familyName}
	easternOrdered := []string{familyName, middleName, givenName}

	join := func(parts []string, sep string) string {
		var nonempty []string
		for _, part := range parts {
			if part != "" {
				nonempty = append(nonempty, part)
			}
		}
		return strings.Join(nonempty, sep)
	}

	if ok := isCJK(lang); ok {
		return join(easternOrdered, "")
	}
	if ok := isEasternNameOrder(lang); ok {
		return join(easternOrdered, " ")
	}

	return join(westernOrdered, " ")
}
