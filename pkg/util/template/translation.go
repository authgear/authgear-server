package template

import (
	"fmt"
	htmltemplate "html/template"
	"strings"
	texttemplate "text/template"
	"text/template/parse"

	messageformat "github.com/iawaknahc/gomessageformat"
)

type TranslationMap struct {
	validator *Validator
	items     map[string]*parse.Tree
}

func (t *TranslationMap) RenderText(key string, args interface{}) (string, error) {
	tree, ok := t.items[key]
	if !ok {
		return "", &errNotFound{name: key}
	}

	template := texttemplate.New("")
	template.Funcs(texttemplate.FuncMap{
		messageformat.TemplateRuntimeFuncName: messageformat.TemplateRuntimeFunc,
		"makemap":                             MakeMap,
	})
	_, err := template.AddParseTree("translation", tree)
	if err != nil {
		return "", fmt.Errorf("template: failed to construct template: %w", err)
	}
	template = template.Lookup("translation")

	err = t.validator.ValidateTextTemplate(template)
	if err != nil {
		return "", fmt.Errorf("template: failed to validate template: %w", err)
	}

	var buf strings.Builder
	err = template.Execute(NewLimitWriter(&buf), args)
	if err != nil {
		return "", fmt.Errorf("template: failed to render template: %w", err)
	}

	return buf.String(), nil
}

func (t *TranslationMap) RenderHTML(key string, args interface{}) (string, error) {
	tree, ok := t.items[key]
	if !ok {
		return "", &errNotFound{name: key}
	}

	template := htmltemplate.New("")
	template.Funcs(htmltemplate.FuncMap{
		messageformat.TemplateRuntimeFuncName: messageformat.TemplateRuntimeFunc,
		"makemap":                             MakeMap,
	})
	_, err := template.AddParseTree("translation", tree)
	if err != nil {
		return "", fmt.Errorf("template: failed to construct template: %w", err)
	}
	template = template.Lookup("translation")

	err = t.validator.ValidateHTMLTemplate(template)
	if err != nil {
		return "", fmt.Errorf("template: failed to validate template: %w", err)
	}

	var buf strings.Builder
	err = template.Execute(NewLimitWriter(&buf), args)
	if err != nil {
		return "", fmt.Errorf("template: failed to render template: %w", err)
	}

	return buf.String(), nil
}
