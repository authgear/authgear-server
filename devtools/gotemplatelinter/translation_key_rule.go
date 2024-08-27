package main

import (
	htmltemplate "html/template"
	"sort"
	"text/template/parse"

	authgearutiltemplate "github.com/authgear/authgear-server/pkg/util/template"
)

type TranslationKeyRule struct{}

func (r TranslationKeyRule) Check(content string, path string) LintViolations {
	return r.check(content, path)
}

func (r TranslationKeyRule) check(content string, path string) LintViolations {
	var violations LintViolations
	t := r.makeTemplate(content)

	r.validateHTMLTemplate(t)

	// TODO: add violations
	return violations
}

func (r TranslationKeyRule) makeTemplate(content string) *htmltemplate.Template {
	t := htmltemplate.New("")
	funcMap := authgearutiltemplate.MakeTemplateFuncMap(t)
	t.Funcs(funcMap)
	parsed := htmltemplate.Must(t.Parse(content))
	return parsed
}

func (r TranslationKeyRule) validateHTMLTemplate(template *htmltemplate.Template) error {
	tpls := template.Templates()

	sort.Slice(tpls, func(i, j int) bool {
		return tpls[i].Name() < tpls[j].Name()
	})

	for _, tpl := range tpls {
		if tpl.Tree == nil {
			continue
		}
		if err := validateTree(tpl.Tree); err != nil {
			return err
		}
	}
	return nil
}

func validateTree(tree *parse.Tree) (err error) {
	validateFn := func(n parse.Node, depth int) (cont bool) {
		switch n := n.(type) {
		case *parse.CommandNode:
			for _, arg := range n.Args {
				if variable, ok := arg.(*parse.VariableNode); ok && variable.String() == "$.Translations.RenderText" {
					// TODO: handle $.Translations.RenderText
				}
				if field, ok := arg.(*parse.FieldNode); ok && field.String() == ".Translations.RenderText" {
					// TODO: handle .Translations.RenderText
				}
				if ident, ok := arg.(*parse.IdentifierNode); ok {
					switch ident.String() {
					case "include":
						// TODO: handle include fn
					case "translate":
						// TODO: handle translate fn
					}
				}
			}
		case *parse.TemplateNode:
			// TODO: handle template node
		}

		return err == nil
	}

	TraverseTree(tree, validateFn)
	return
}
