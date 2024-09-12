package main

import (
	"text/template/parse"
)

// validate commands with variable `$.Translations.HasKey` or field `.Translations.HasKey`
//
// example: `($.Translations.HasKey "customer-support-link" nil)`
// example: (.Translations.HasKey "terms-of-service-link" nil)
func CheckCommandTranslationsHasKey(node *parse.CommandNode) (err error) {
	// 2nd arg should be translation key
	for idx, arg := range node.Args {
		if idx == 1 {
			err = CheckTranslationKeyNode(arg)
			if err != nil {
				return err
			}
		}

	}
	return
}
