package template

import (
	"fmt"
	htmltemplate "html/template"
	"strings"
	texttemplate "text/template"
	"text/template/parse"

	"github.com/iawaknahc/gomessageformat"
	"golang.org/x/text/language"
)

//go:generate mockgen -source=engine.go -destination=engine_mock_test.go -package template

// nolint:golint
type TemplateResolver interface {
	Resolve(ctx *ResolveContext, typ string) (*Resolved, error)
	ResolveTranslations(ctx *ResolveContext, typ string) (map[string]Translation, error)
}

type RenderContext struct {
	ValidatorOptions      []ValidatorOption
	PreferredLanguageTags []string
}

type Engine struct {
	Resolver TemplateResolver
}

func (e *Engine) RenderTranslation(ctx *RenderContext, typ string, key string, data interface{}) (string, error) {
	translations, err := e.Resolver.ResolveTranslations(
		&ResolveContext{PreferredLanguageTags: ctx.PreferredLanguageTags},
		typ,
	)
	if err != nil {
		return "", err
	}

	return e.renderText(ctx, &Resolved{
		T:                 T{},
		Content:           fmt.Sprintf("{{ template %q . }}", key),
		Translations:      translations,
		ComponentContents: nil,
	}, data)
}

func (e *Engine) Render(ctx *RenderContext, typ string, data interface{}) (out string, err error) {
	resolveCtx := &ResolveContext{
		PreferredLanguageTags: ctx.PreferredLanguageTags,
	}

	resolved, err := e.Resolver.Resolve(resolveCtx, typ)
	if err != nil {
		return
	}

	if resolved.T.IsHTML {
		return e.renderHTML(ctx, resolved, data)
	}

	return e.renderText(ctx, resolved, data)
}

func (e *Engine) renderHTML(ctx *RenderContext, resolved *Resolved, data interface{}) (out string, err error) {
	t := htmltemplate.New(resolved.T.Type)

	// Inject the funcs map before parsing any templates.
	// This is required by the documentation.
	t.Funcs(htmltemplate.FuncMap{
		messageformat.TemplateRuntimeFuncName: messageformat.TemplateRuntimeFunc,
		"makemap":                             MakeMap,
	})

	// Parse the main template.
	_, err = t.Parse(resolved.Content)
	if err != nil {
		err = fmt.Errorf("template: failed to parse main template content: %w", err)
		return
	}

	// Parse Defines.
	for _, define := range resolved.T.Defines {
		_, err = t.Parse(define)
		if err != nil {
			err = fmt.Errorf("template: failed to parse template define: %w", err)
			return
		}
	}

	// Parse components.
	for _, component := range resolved.ComponentContents {
		_, err = t.Parse(component)
		if err != nil {
			err = fmt.Errorf("template: failed to parse template component: %w", err)
			return
		}
	}

	// Parse translations.
	for key, translation := range resolved.Translations {
		var tree *parse.Tree
		tag := language.Make(translation.LanguageTag)
		tree, err = messageformat.FormatTemplateParseTree(tag, translation.Value)
		if err != nil {
			err = fmt.Errorf("template: failed to parse messageformat: %w", err)
			return
		}

		_, err = t.AddParseTree(key, tree)
		if err != nil {
			err = fmt.Errorf("template: failed to add messageformat parse tree: %w", err)
			return
		}
	}

	// Validate all templates
	validator := NewValidator(ctx.ValidatorOptions...)
	err = validator.ValidateHTMLTemplate(t)
	if err != nil {
		err = fmt.Errorf("template: failed to validate html template: %w", err)
		return
	}

	var buf strings.Builder
	err = t.Execute(NewLimitWriter(&buf), data)
	if err != nil {
		err = fmt.Errorf("template: failed to execute html template: %w", err)
		return
	}

	out = buf.String()
	return
}

func (e *Engine) renderText(ctx *RenderContext, resolved *Resolved, data interface{}) (out string, err error) {
	t := texttemplate.New(resolved.T.Type)

	// Inject the funcs map before parsing any templates.
	// This is required by the documentation.
	t.Funcs(texttemplate.FuncMap{
		messageformat.TemplateRuntimeFuncName: messageformat.TemplateRuntimeFunc,
		"makemap": func(pairs ...interface{}) map[string]interface{} {
			out := make(map[string]interface{})
			for i := 0; i < len(pairs); i += 2 {
				key := pairs[i].(string)
				value := pairs[i+1]
				out[key] = value
			}
			return out
		},
	})

	// Parse the main template.
	_, err = t.Parse(resolved.Content)
	if err != nil {
		err = fmt.Errorf("template: failed to parse main template content: %w", err)
		return
	}

	// Parse Defines.
	for _, define := range resolved.T.Defines {
		_, err = t.Parse(define)
		if err != nil {
			err = fmt.Errorf("template: failed to parse template define: %w", err)
			return
		}
	}

	// Parse components.
	for _, component := range resolved.ComponentContents {
		_, err = t.Parse(component)
		if err != nil {
			err = fmt.Errorf("template: failed to parse template component: %w", err)
			return
		}
	}

	// Parse translations.
	for key, translation := range resolved.Translations {
		var tree *parse.Tree
		tag := language.Make(translation.LanguageTag)
		tree, err = messageformat.FormatTemplateParseTree(tag, translation.Value)
		if err != nil {
			err = fmt.Errorf("template: failed to parse messageformat: %w", err)
			return
		}

		_, err = t.AddParseTree(key, tree)
		if err != nil {
			err = fmt.Errorf("template: failed to add messageformat parse tree: %w", err)
			return
		}
	}

	// Validate all templates
	validator := NewValidator(ctx.ValidatorOptions...)
	err = validator.ValidateTextTemplate(t)
	if err != nil {
		err = fmt.Errorf("template: failed to validate html template: %w", err)
		return
	}

	var buf strings.Builder
	err = t.Execute(NewLimitWriter(&buf), data)
	if err != nil {
		err = fmt.Errorf("template: failed to execute html template: %w", err)
		return
	}

	out = buf.String()
	return
}
