package intl

import (
	"golang.org/x/text/language"
)

func HTMLDir(languageTag string) string {
	tag := language.Make(languageTag)
	base, _ := tag.Base()

	rtlScript, ok := rtlMap[base.String()]
	if !ok {
		// No present in the rtl map, then it must be ltr.
		return "ltr"
	}

	if rtlScript == "" {
		// Not script-specific, so the entire language is rtl.
		return "rtl"
	}

	script, _ := tag.Script()
	if rtlScript == script.String() {
		// This language is not rtl, but when this script is used, then it is rtl.
		return "rtl"
	}

	// Otherwise it is ltr.
	return "ltr"
}
