package config

import (
	"fmt"
	"text/template"
)

var _ = Schema.Add("TextTemplate", `
{
	"type": "object",
	"required": ["text_template"],
	"additionalProperties": false,
	"properties": {
		"text_template": {
			"type": "object",
			"required": ["template"],
			"additionalProperties": false,
			"properties": {
				"template": { "type": "string", "format": "x_text_template" }
			}
		}
	}
}
`)

type TextTemplate struct {
	TextTemplate *TextTemplateBody `json:"text_template,omitempty" nullable:"true"`
}

func (t *TextTemplate) MustGetTextTemplate() *template.Template {
	tpl := template.New("")
	var err error
	tpl, err = tpl.Parse(t.TextTemplate.Template)
	if err != nil {
		panic(fmt.Errorf("cannot parse template in config: %w", err))
	}
	return tpl
}

type TextTemplateBody struct {
	Template string `json:"template,omitempty"`
}
