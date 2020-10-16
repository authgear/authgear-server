package template

import (
	"fmt"
	htmltemplate "html/template"
	"strings"
	texttemplate "text/template"
	"text/template/parse"

	messageformat "github.com/iawaknahc/gomessageformat"
	"golang.org/x/text/language"
)

type EngineTemplateResolver interface {
	ResolveHTML(desc *HTML, preferredLanguages []string) (*htmltemplate.Template, error)
	ResolvePlainText(desc *PlainText, preferredLanguages []string) (*texttemplate.Template, error)
	ResolveTranslations(preferredLanguages []string) (map[string]Translation, error)
}

type Engine struct {
	Resolver EngineTemplateResolver
}

func (e *Engine) Translation(preferredLanguages []string) (*TranslationMap, error) {
	translations, err := e.Resolver.ResolveTranslations(preferredLanguages)
	if err != nil {
		return nil, err
	}

	// Parse translations.
	var items = make(map[string]*parse.Tree)
	for key, translation := range translations {
		var tree *parse.Tree
		tag := language.Make(translation.LanguageTag)
		tree, err = messageformat.FormatTemplateParseTree(tag, translation.Value)
		if err != nil {
			return nil, fmt.Errorf("failed to parse messageformat: %w", err)
		}

		items[key] = tree
	}

	return &TranslationMap{items: items, validator: templateValidator}, nil
}

func (e *Engine) Render(resource Resource, preferredLanguages []string, data interface{}) (string, error) {
	switch desc := resource.(type) {
	case *HTML:
		return e.renderHTML(desc, preferredLanguages, data)
	case *PlainText:
		return e.renderPlainText(desc, preferredLanguages, data)
	default:
		panic("template: unexpected template resource type")
	}
}

func (e *Engine) renderHTML(desc *HTML, preferredLanguages []string, data interface{}) (string, error) {
	t := htmltemplate.New("")
	t.Funcs(templateFuncMap)

	var loadTemplate func(desc *HTML) error
	loadTemplate = func(desc *HTML) error {
		// Include main template.
		tpl, err := e.Resolver.ResolveHTML(desc, preferredLanguages)
		if err != nil {
			return fmt.Errorf("failed to load template: %w", err)
		}
		for _, tpl := range tpl.Templates() {
			if _, err := t.AddParseTree(tpl.Name(), tpl.Tree); err != nil {
				return fmt.Errorf("failed to add template parse tree: %w", err)
			}
		}

		// Include component dependencies.
		for _, component := range desc.ComponentDependencies {
			if err := loadTemplate(component); err != nil {
				return err
			}
		}

		return nil
	}
	if err := loadTemplate(desc); err != nil {
		return "", err
	}
	t = t.Lookup(desc.Name)

	// Include translations.
	translations, err := e.Resolver.ResolveTranslations(preferredLanguages)
	if err != nil {
		return "", fmt.Errorf("failed to load translation: %w", err)
	}
	for key, translation := range translations {
		tag := language.Make(translation.LanguageTag)
		tree, err := messageformat.FormatTemplateParseTree(tag, translation.Value)
		if err != nil {
			return "", fmt.Errorf("failed to parse messageformat: %w", err)
		}

		_, err = t.AddParseTree(key, tree)
		if err != nil {
			return "", fmt.Errorf("failed to add messageformat parse tree: %w", err)
		}
	}

	// Validate all templates
	err = templateValidator.ValidateHTMLTemplate(t)
	if err != nil {
		return "", fmt.Errorf("invalid html template: %w", err)
	}

	var buf strings.Builder
	err = t.Execute(NewLimitWriter(&buf), data)
	if err != nil {
		return "", fmt.Errorf("failed to execute html template: %w", err)
	}

	return buf.String(), nil
}

func (e *Engine) renderPlainText(desc *PlainText, preferredLanguages []string, data interface{}) (string, error) {
	t := texttemplate.New("")
	t.Funcs(templateFuncMap)

	var loadTemplate func(desc *PlainText) error
	loadTemplate = func(desc *PlainText) error {
		// Include main template.
		tpl, err := e.Resolver.ResolvePlainText(desc, preferredLanguages)
		if err != nil {
			return fmt.Errorf("failed to load template: %w", err)
		}
		for _, tpl := range tpl.Templates() {
			if _, err := t.AddParseTree(tpl.Name(), tpl.Tree); err != nil {
				return fmt.Errorf("failed to add template parse tree: %w", err)
			}
		}

		// Include component dependencies.
		for _, component := range desc.ComponentDependencies {
			if err := loadTemplate(component); err != nil {
				return err
			}
		}

		return nil
	}
	if err := loadTemplate(desc); err != nil {
		return "", err
	}
	t = t.Lookup(desc.Name)

	// Include translations.
	translations, err := e.Resolver.ResolveTranslations(preferredLanguages)
	if err != nil {
		return "", fmt.Errorf("failed to load translation: %w", err)
	}
	for key, translation := range translations {
		tag := language.Make(translation.LanguageTag)
		tree, err := messageformat.FormatTemplateParseTree(tag, translation.Value)
		if err != nil {
			return "", fmt.Errorf("failed to parse messageformat: %w", err)
		}

		_, err = t.AddParseTree(key, tree)
		if err != nil {
			return "", fmt.Errorf("failed to add messageformat parse tree: %w", err)
		}
	}

	// Validate all templates
	err = templateValidator.ValidateTextTemplate(t)
	if err != nil {
		return "", fmt.Errorf("invalid text template: %w", err)
	}

	var buf strings.Builder
	err = t.Execute(NewLimitWriter(&buf), data)
	if err != nil {
		return "", fmt.Errorf("failed to execute text template: %w", err)
	}

	return buf.String(), nil
}
