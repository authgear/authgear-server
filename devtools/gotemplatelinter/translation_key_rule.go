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
	t := r.makeTemplate(content)

	violations := r.validateHTMLTemplate(t, path)
	return violations
}

func (r TranslationKeyRule) makeTemplate(content string) *htmltemplate.Template {
	t := htmltemplate.New("")
	funcMap := authgearutiltemplate.MakeTemplateFuncMap(t)
	t.Funcs(funcMap)
	parsed := htmltemplate.Must(t.Parse(content))
	return parsed
}

func (r TranslationKeyRule) validateHTMLTemplate(template *htmltemplate.Template, path string) LintViolations {
	tpls := template.Templates()

	var violations LintViolations
	sort.Slice(tpls, func(i, j int) bool {
		return tpls[i].Name() < tpls[j].Name()
	})

	for _, tpl := range tpls {
		if tpl.Tree == nil {
			continue
		}
		if tplViolations := validateTree(tpl.Tree, path); len(tplViolations) != 0 {
			violations = append(violations, tplViolations...)
		}
	}
	return violations
}

func validateTree(tree *parse.Tree, path string) LintViolations {
	var violations LintViolations
	var err error

	handleNodeErrFn := func(n parse.Node, _err error) {
		if _err == nil {
			return
		}
		line, col, _ := TreeErrorContext(tree, n)
		violations = append(violations, LintViolation{
			Line:    line,
			Column:  col,
			Path:    path,
			Message: _err.Error(),
		})
	}

	validateFn := func(n parse.Node, depth int) (cont bool) {
		switch n := n.(type) {
		case *parse.CommandNode:
			for _, arg := range n.Args {
				if variable, ok := arg.(*parse.VariableNode); ok && variable.String() == "$.Translations.RenderText" {
					err = CheckCommandTranslationsRenderText(n)
					handleNodeErrFn(n, err)
				}
				if field, ok := arg.(*parse.FieldNode); ok && field.String() == ".Translations.RenderText" {
					err = CheckCommandTranslationsRenderText(n)
					handleNodeErrFn(n, err)
				}
				if ident, ok := arg.(*parse.IdentifierNode); ok {
					switch ident.String() {
					case "include":
						err = CheckCommandInclude(n)
						handleNodeErrFn(n, err)
					case "translate":
						err = CheckCommandTranslate(n)
						handleNodeErrFn(n, err)
					}
				}
			}
		case *parse.TemplateNode:
			err = CheckTemplate(n)
			handleNodeErrFn(n, err)
		}

		// always continue to traverse
		return true
	}

	TraverseTree(tree, validateFn)
	return violations
}
