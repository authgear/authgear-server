package main

import (
	"strings"
	"text/template/parse"
)

// validate `include` command
//
// e.g. (include "v2-error-screen-title" nil)
// e.g. (include (printf "v2-oauth-branding-%s" .provider_type) nil)
// e.g. (include $description_key nil)
func CheckCommandInclude(includeNode *parse.CommandNode) (err error) {
	// 2nd arg should be translation key
	for idx, arg := range includeNode.Args {
		if idx == 1 {
			if isHTMLInclude(arg) { // false positive
				break
			}
			err = CheckTranslationKeyNode(arg)
			if err != nil {
				return err
			}
		}

	}
	return
}

func isHTMLInclude(arg parse.Node) bool {
	if str, ok := arg.(*parse.StringNode); ok {
		if strings.Contains(str.Text, ".html") {
			return true
		}
	}
	return false
}
