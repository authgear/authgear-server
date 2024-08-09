package e2e

import (
	"text/template"

	"github.com/Masterminds/sprig"
)

func ParseSQLTemplate(name string, rawTemplateText string) (*template.Template, error) {
	tmpl := template.New(name)
	tmpl.Funcs(sprig.GenericFuncMap())
	sqlTmpl, err := tmpl.Parse(rawTemplateText)
	return sqlTmpl, err
}
