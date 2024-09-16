package template

import (
	htmltemplate "html/template"
	texttemplate "text/template"
)

type HTMLTemplateEffectiveResource struct {
	Data        []byte
	LanguageTag string
	Template    *htmltemplate.Template
}

type TextTemplateEffectiveResource struct {
	Data        []byte
	LanguageTag string
	Template    *texttemplate.Template
}
