package template

import (
	"fmt"
	htmltemplate "html/template"
	"net/http"
	"strconv"
	"strings"
	texttemplate "text/template"
	"text/template/parse"

	"golang.org/x/text/language"

	"github.com/authgear/authgear-server/pkg/util/intl"
	"github.com/authgear/authgear-server/pkg/util/messageformat"
)

type RenderResult struct {
	String      string
	LanguageTag string
}

type EngineTemplateResolver interface {
	ResolveHTML(desc *HTML, preferredLanguages []string) (*HTMLTemplateEffectiveResource, error)
	ResolveMessageHTML(desc *MessageHTML, preferredLanguages []string) (*HTMLTemplateEffectiveResource, error)
	ResolvePlainText(desc *PlainText, preferredLanguages []string) (*TextTemplateEffectiveResource, error)
	ResolveMessagePlainText(desc *MessagePlainText, preferredLanguages []string) (*TextTemplateEffectiveResource, error)
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
			return nil, fmt.Errorf("failed to parse messageformat for key %s: %w", key, err)
		}

		items[key] = tree
	}

	return &TranslationMap{items: items}, nil
}

func (e *Engine) Render(resource Resource, preferredLanguages []string, data interface{}) (*RenderResult, error) {
	switch desc := resource.(type) {
	case *HTML:
		return e.renderHTML(desc, preferredLanguages, data)
	case *MessageHTML:
		return e.renderMessageHTML(desc, preferredLanguages, data)
	case *PlainText:
		return e.renderPlainText(desc, preferredLanguages, data)
	case *MessagePlainText:
		return e.renderMessagePlainText(desc, preferredLanguages, data)
	default:
		panic("template: unexpected template resource type")
	}
}

func (e *Engine) renderHTML(desc *HTML, preferredLanguages []string, data interface{}) (*RenderResult, error) {
	t := htmltemplate.New("")
	funcMap := MakeTemplateFuncMap(t)
	t.Funcs(funcMap)

	var loadTemplate func(desc *HTML) (string, error)
	loadTemplate = func(desc *HTML) (string, error) {
		// Include main template.
		h, err := e.Resolver.ResolveHTML(desc, preferredLanguages)
		tpl := h.Template
		if err != nil {
			return "", fmt.Errorf("failed to load template %s: %w", desc.Name, err)
		}
		for _, tpl := range tpl.Templates() {
			if _, err := t.AddParseTree(tpl.Name(), tpl.Tree); err != nil {
				return "", fmt.Errorf("failed to add template parse tree: %w", err)
			}
		}

		// Include component dependencies.
		for _, component := range desc.ComponentDependencies {
			if _, err := loadTemplate(component); err != nil {
				return "", err
			}
		}

		return h.LanguageTag, nil
	}
	languageTag, err := loadTemplate(desc)
	if err != nil {
		return nil, err
	}
	t = t.Lookup(desc.Name)

	// Include translations.
	translations, err := e.Resolver.ResolveTranslations(preferredLanguages)
	if err != nil {
		return nil, fmt.Errorf("failed to load translation: %w", err)
	}
	for key, translation := range translations {
		tag := language.Make(translation.LanguageTag)
		tree, err := messageformat.FormatTemplateParseTree(tag, translation.Value)
		if err != nil {
			return nil, fmt.Errorf("failed to parse messageformat for key %s: %w", key, err)
		}

		_, err = t.AddParseTree(key, tree)
		if err != nil {
			return nil, fmt.Errorf("failed to add messageformat parse tree for key %s: %w", key, err)
		}
	}

	var buf strings.Builder
	err = t.Execute(NewLimitWriter(&buf), data)
	if err != nil {
		return nil, fmt.Errorf("failed to execute html template %s: %w", desc.Name, err)
	}

	return &RenderResult{String: buf.String(), LanguageTag: string(languageTag)}, nil
}

func (e *Engine) renderMessageHTML(desc *MessageHTML, preferredLanguages []string, data interface{}) (*RenderResult, error) {
	t := htmltemplate.New("")
	funcMap := MakeTemplateFuncMap(t)
	t.Funcs(funcMap)

	var loadTemplate func(desc *MessageHTML) (string, error)
	loadTemplate = func(desc *MessageHTML) (string, error) {
		// Include main template.
		h, err := e.Resolver.ResolveMessageHTML(desc, preferredLanguages)
		tpl := h.Template
		if err != nil {
			return "", fmt.Errorf("failed to load template %s: %w", desc.Name, err)
		}
		for _, tpl := range tpl.Templates() {
			if _, err := t.AddParseTree(tpl.Name(), tpl.Tree); err != nil {
				return "", fmt.Errorf("failed to add template parse tree: %w", err)
			}
		}

		// No component dependencies for message html

		return h.LanguageTag, nil
	}
	languageTag, err := loadTemplate(desc)
	if err != nil {
		return nil, err
	}
	t = t.Lookup(desc.Name)

	// Include translations.
	translations, err := e.Resolver.ResolveTranslations(preferredLanguages)
	if err != nil {
		return nil, fmt.Errorf("failed to load translation: %w", err)
	}
	for key, translation := range translations {
		tag := language.Make(translation.LanguageTag)
		tree, err := messageformat.FormatTemplateParseTree(tag, translation.Value)
		if err != nil {
			return nil, fmt.Errorf("failed to parse messageformat for key %s: %w", key, err)
		}

		_, err = t.AddParseTree(key, tree)
		if err != nil {
			return nil, fmt.Errorf("failed to add messageformat parse tree for key %s: %w", key, err)
		}
	}

	var buf strings.Builder
	err = t.Execute(NewLimitWriter(&buf), data)
	if err != nil {
		return nil, fmt.Errorf("failed to execute html template %s: %w", desc.Name, err)
	}

	return &RenderResult{String: buf.String(), LanguageTag: string(languageTag)}, nil
}

func (e *Engine) renderPlainText(desc *PlainText, preferredLanguages []string, data interface{}) (*RenderResult, error) {
	t := texttemplate.New("")
	funcMap := MakeTemplateFuncMap(t)
	t.Funcs(funcMap)

	var loadTemplate func(desc *PlainText) (string, error)
	loadTemplate = func(desc *PlainText) (string, error) {
		// Include main template.
		h, err := e.Resolver.ResolvePlainText(desc, preferredLanguages)
		tpl := h.Template
		if err != nil {
			return "", fmt.Errorf("failed to load template: %w", err)
		}
		for _, tpl := range tpl.Templates() {
			if _, err := t.AddParseTree(tpl.Name(), tpl.Tree); err != nil {
				return "", fmt.Errorf("failed to add template parse tree: %w", err)
			}
		}

		// Include component dependencies.
		for _, component := range desc.ComponentDependencies {
			if _, err := loadTemplate(component); err != nil {
				return "", err
			}
		}

		return h.LanguageTag, nil
	}
	languageTag, err := loadTemplate(desc)
	if err != nil {
		return nil, err
	}
	t = t.Lookup(desc.Name)

	// Include translations.
	translations, err := e.Resolver.ResolveTranslations(preferredLanguages)
	if err != nil {
		return nil, fmt.Errorf("failed to load translation: %w", err)
	}
	for key, translation := range translations {
		tag := language.Make(translation.LanguageTag)
		tree, err := messageformat.FormatTemplateParseTree(tag, translation.Value)
		if err != nil {
			return nil, fmt.Errorf("failed to parse messageformat for key %s: %w", key, err)
		}

		_, err = t.AddParseTree(key, tree)
		if err != nil {
			return nil, fmt.Errorf("failed to add messageformat parse tree for key %s: %w", key, err)
		}
	}

	var buf strings.Builder
	err = t.Execute(NewLimitWriter(&buf), data)
	if err != nil {
		return nil, fmt.Errorf("failed to execute text template %s: %w", desc.Name, err)
	}

	return &RenderResult{String: buf.String(), LanguageTag: string(languageTag)}, nil
}

func (e *Engine) renderMessagePlainText(desc *MessagePlainText, preferredLanguages []string, data interface{}) (*RenderResult, error) {
	t := texttemplate.New("")
	funcMap := MakeTemplateFuncMap(t)
	t.Funcs(funcMap)

	var loadTemplate func(desc *MessagePlainText) (string, error)
	loadTemplate = func(desc *MessagePlainText) (string, error) {
		// Include main template.
		h, err := e.Resolver.ResolveMessagePlainText(desc, preferredLanguages)
		tpl := h.Template
		if err != nil {
			return "", fmt.Errorf("failed to load template: %w", err)
		}
		for _, tpl := range tpl.Templates() {
			if _, err := t.AddParseTree(tpl.Name(), tpl.Tree); err != nil {
				return "", fmt.Errorf("failed to add template parse tree: %w", err)
			}
		}

		// No component dependencies for message plain text

		return h.LanguageTag, nil
	}
	languageTag, err := loadTemplate(desc)
	if err != nil {
		return nil, err
	}
	t = t.Lookup(desc.Name)

	// Include translations.
	translations, err := e.Resolver.ResolveTranslations(preferredLanguages)
	if err != nil {
		return nil, fmt.Errorf("failed to load translation: %w", err)
	}
	for key, translation := range translations {
		tag := language.Make(translation.LanguageTag)
		tree, err := messageformat.FormatTemplateParseTree(tag, translation.Value)
		if err != nil {
			return nil, fmt.Errorf("failed to parse messageformat for key %s: %w", key, err)
		}

		_, err = t.AddParseTree(key, tree)
		if err != nil {
			return nil, fmt.Errorf("failed to add messageformat parse tree for key %s: %w", key, err)
		}
	}

	var buf strings.Builder
	err = t.Execute(NewLimitWriter(&buf), data)
	if err != nil {
		return nil, fmt.Errorf("failed to execute text template %s: %w", desc.Name, err)
	}

	return &RenderResult{String: buf.String(), LanguageTag: string(languageTag)}, nil
}

func (e *Engine) RenderStatus(w http.ResponseWriter, r *http.Request, status int, tpl Resource, data interface{}) {
	preferredLanguageTags := intl.GetPreferredLanguageTags(r.Context())
	out, err := e.Render(
		tpl,
		preferredLanguageTags,
		data,
	)
	if err != nil {
		panic(err)
	}

	body := []byte(out.String)
	w.Header().Set("Content-Length", strconv.Itoa(len(body)))
	w.WriteHeader(status)
	_, err = w.Write(body)
	if err != nil {
		panic(err)
	}
}
