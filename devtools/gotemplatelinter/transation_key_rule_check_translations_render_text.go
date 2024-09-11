package main

import (
	"fmt"
	"text/template/parse"
)

// validate commands with variable `$.Translations.RenderText` or field `.Translations.RenderText`
//
// example: `($.Translations.RenderText "customer-support-link" nil)`
// example: (.Translations.RenderText "terms-of-service-link" nil)
func CheckCommandTranslationsRenderText(node *parse.CommandNode) (err error) {
	return fmt.Errorf("Translations.RenderText is forbidden: `%s`", node.String())
}
