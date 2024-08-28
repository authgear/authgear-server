package main

import (
	"fmt"
	"strings"
	"text/template/parse"
)

// validate `{{ template }}` nodes
//
// e.g. {{template "setup-totp-get-google-authenticator-description"}}
// e.g. {{template "setup-totp-raw-secret" (dict "secret" $.Secret)}}
// e.g. {{template "settings-totp-item-description" (dict "time" .CreatedAt "rfc3339" (rfc3339 .CreatedAt))}}
func CheckTemplate(templateNode *parse.TemplateNode) (err error) {
	if isHTMLTemplate(templateNode) { // false positive
		return
	}
	translationKey := templateNode.Name
	err = CheckTranslationKey(translationKey)
	if err != nil {
		return fmt.Errorf("invalid template name: %w", err)
	}
	return
}

func isHTMLTemplate(templateNode *parse.TemplateNode) bool {
	return strings.Contains(templateNode.Name, ".html")
}
