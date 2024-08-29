package main

import (
	"fmt"
	"regexp"
	"strings"
	"text/template/parse"
)

const TranslationKeyPattern = `^(v2)\.(page|component)\.([-a-z]+)\.(default)\.([-a-z]+)$`

var validKey *regexp.Regexp

func init() {
	validKey = regexp.MustCompile(TranslationKeyPattern)
}

func CheckTranslationKeyNode(translationKeyNode parse.Node) (err error) {
	switch translationKeyNode.Type() {
	case parse.NodeString:
		return CheckTranslationKeyPattern(translationKeyNode.(*parse.StringNode).Text)
	case parse.NodeVariable:
		// FIXME: support variable, like $label, $label_key, $variant_label_key
		fallthrough
	case parse.NodePipe:
		// FIXME: support pipe, like (printf "territory-%s" $.AddressCountry)
		fallthrough
	default:
		return fmt.Errorf("invalid translation key: \"%v\"", translationKeyNode.String())
	}
}

func CheckTranslationKeyPattern(translationKey string) (err error) {
	key := strings.Trim(translationKey, "\"")

	if key == "" {
		return fmt.Errorf("translation key is empty")
	}

	if ok := validKey.MatchString(key); !ok {
		return fmt.Errorf("invalid translation key: \"%v\"", key)
	}

	return
}
