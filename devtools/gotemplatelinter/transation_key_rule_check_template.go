package main

import (
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
	err = CheckTranslationKeyPattern(translationKey)
	if err != nil {
		return err
	}
	return
}

func isHTMLTemplate(templateNode *parse.TemplateNode) bool {
	return strings.Contains(templateNode.Name, ".html")
}
