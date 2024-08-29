package main

import (
	"fmt"
	"regexp"
	"strings"
	"text/template/parse"
)

func CheckTranslationKeyNode(translationKeyNode parse.Node) (err error) {
	switch translationKeyNode.Type() {
	case parse.NodeString:
		return CheckTranslationKey(translationKeyNode.(*parse.StringNode).Text)
	case parse.NodeVariable:
		// FIXME: support variable, like $label, $label_key, $variant_label_key
		fallthrough
	case parse.NodePipe:
		// FIXME: support pipe, like (printf "territory-%s" $.AddressCountry)
		fallthrough
	default:
		return fmt.Errorf("expected: *parse.StringNode, got: %T \n\t\t concerning node: %v", translationKeyNode, translationKeyNode)
	}
}

const TranslationKeyPattern = `(v2)\.(page|component)\.([-a-z]+)\.(default)\.([-a-z]+)`

func CheckTranslationKey(translationKey string) (err error) {
	key := strings.Trim(translationKey, "\"")

	if key == "" {
		return fmt.Errorf("translation key is empty")
	}

	var validKey = regexp.MustCompile(TranslationKeyPattern)
	if ok := validKey.MatchString(key); !ok {
		return fmt.Errorf("invalid translation key: \"%v\"", key)
	}

	return
}
