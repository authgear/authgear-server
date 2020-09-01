package template

import (
	"fmt"
	"strings"
	texttemplate "text/template"
)

type TranslationMap struct {
	template *texttemplate.Template
}

func (t *TranslationMap) Render(key string, args interface{}) (string, error) {
	template := t.template.Lookup(key)
	if template == nil {
		return "", &errNotFound{name: key}
	}

	var buf strings.Builder
	err := template.Execute(NewLimitWriter(&buf), args)
	if err != nil {
		return "", fmt.Errorf("template: failed to render template: %w", err)
	}

	return buf.String(), nil
}
