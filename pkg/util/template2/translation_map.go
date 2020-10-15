package template

import (
	"fmt"
	htmltemplate "html/template"
	"strings"
	texttemplate "text/template"
	"text/template/parse"

	"github.com/authgear/authgear-server/pkg/util/template"
)

type TranslationMap struct {
	validator *template.Validator
	items     map[string]*parse.Tree
}

func (t *TranslationMap) RenderText(key string, args interface{}) (string, error) {
	tree, ok := t.items[key]
	if !ok {
		return "", fmt.Errorf("translation key not found: %s", key)
	}

	tpl := texttemplate.New("")
	tpl.Funcs(templateFuncMap)
	_, err := tpl.AddParseTree("translation", tree)
	if err != nil {
		return "", fmt.Errorf("template: failed to construct template: %w", err)
	}
	tpl = tpl.Lookup("translation")

	err = t.validator.ValidateTextTemplate(tpl)
	if err != nil {
		return "", fmt.Errorf("template: failed to validate template: %w", err)
	}

	var buf strings.Builder
	err = tpl.Execute(template.NewLimitWriter(&buf), args)
	if err != nil {
		return "", fmt.Errorf("template: failed to render template: %w", err)
	}

	return buf.String(), nil
}

func (t *TranslationMap) RenderHTML(key string, args interface{}) (string, error) {
	tree, ok := t.items[key]
	if !ok {
		return "", fmt.Errorf("translation key not found: %s", key)
	}

	tpl := htmltemplate.New("")
	tpl.Funcs(templateFuncMap)
	_, err := tpl.AddParseTree("translation", tree)
	if err != nil {
		return "", fmt.Errorf("template: failed to construct template: %w", err)
	}
	tpl = tpl.Lookup("translation")

	err = t.validator.ValidateHTMLTemplate(tpl)
	if err != nil {
		return "", fmt.Errorf("template: failed to validate template: %w", err)
	}

	var buf strings.Builder
	err = tpl.Execute(template.NewLimitWriter(&buf), args)
	if err != nil {
		return "", fmt.Errorf("template: failed to render template: %w", err)
	}

	return buf.String(), nil
}
