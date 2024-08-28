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
	default:
		return fmt.Errorf("expected: *parse.StringNode, got: %T", translationKeyNode)
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
		return fmt.Errorf("invalid translation key, please follow format: `v2.page.my-page.default.my-descriptor`")
	}

	return
}
