package main

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"
	"text/template/parse"
)

// v2.page.<page>.<state>.<descriptor>
// v2.component.<component>.<state>.<descriptor>
const TranslationKeyPattern = `^(v2)\.(page|component)\.([-a-z0-9]+)\.([-a-z0-9]+)\.([-a-z0-9]+)$`
const ErrTranslationKeyPattern = `^(v2)\.(error)\.([-a-z0-9]+)$`
const enTranslationJSONPath = "resources/authgear/templates/en/translation.json"

var validKey *regexp.Regexp
var validErrKey *regexp.Regexp
var translationKeys map[string]struct{}

func init() {
	validKey = regexp.MustCompile(TranslationKeyPattern)
	validErrKey = regexp.MustCompile(ErrTranslationKeyPattern)
	translationKeys = getEnJSONTranslationKeys()
}

func CheckTranslationKeyNode(translationKeyNode parse.Node) (err error) {
	switch translationKeyNode.Type() {
	case parse.NodeString:
		return CheckTranslationKeyPattern(translationKeyNode.(*parse.StringNode).Text)
	case parse.NodeVariable:
		// FIXME: support variable, like $label, $label_key, $variant_label_key
		fallthrough
	case parse.NodePipe:
		// we can skip printf-only pipe nodes here, since it is already checked in CheckCommandPrintf
		if IsPipeNodeOnlyPrintfCommand(translationKeyNode.(*parse.PipeNode)) {
			return
		}
		fallthrough
	default:
		if IsSpecialCase(translationKeyNode.String()) {
			return nil
		}
		return fmt.Errorf("invalid translation key: \"%v\"", translationKeyNode.String())
	}
}

func CheckTranslationKeyPattern(translationKey string) (err error) {
	key := strings.Trim(translationKey, "\"")

	if IsSpecialCase(translationKey) {
		return nil
	}

	if key == "" {
		return fmt.Errorf("translation key is empty")
	}

	if !isTranslationKeyDefined(key) {
		return fmt.Errorf("translation key not defined: \"%v\"", key)
	}

	if !validKey.MatchString(key) && !validErrKey.MatchString(key) {
		return fmt.Errorf("invalid translation key: \"%v\"", key)
	}

	return
}

func isTranslationKeyDefined(targetKey string) bool {
	_, ok := translationKeys[targetKey]
	return ok
}

func getEnJSONTranslationKeys() map[string]struct{} {
	bytes, err := os.ReadFile(enTranslationJSONPath)
	if err != nil {
		panic(fmt.Errorf("failed to read %v: %w", enTranslationJSONPath, err))
	}

	var translationData map[string]string

	err = json.Unmarshal(bytes, &translationData)
	if err != nil {
		panic(fmt.Errorf("failed to unmarshal %v: %w", enTranslationJSONPath, err))
	}

	keys := make(map[string]struct{})
	for key := range translationData {
		keys[key] = struct{}{}
	}
	return keys

}
