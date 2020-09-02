package template

import (
	"encoding/json"
	"fmt"

	"github.com/authgear/authgear-server/pkg/util/fs"
	"github.com/authgear/authgear-server/pkg/util/intl"
)

type Loader interface {
	Load(uri string) (content string, err error)
}

type NewResolverOptions struct {
	AppFs                     fs.Fs
	Registry                  *Registry
	DefaultTemplatesDirectory string
	References                []Reference
	FallbackLanguageTag       string
}

type ResolveContext struct {
	PreferredLanguageTags []string
}

// Resolver resolves templates.
type Resolver struct {
	Loader              Loader
	DefaultLoader       DefaultLoader
	Registry            *Registry
	References          []Reference
	FallbackLanguageTag string
}

func NewResolver(opts NewResolverOptions) *Resolver {
	uriLoader := NewURILoader(opts.AppFs)
	defaultLoader := &DefaultLoaderFS{Directory: opts.DefaultTemplatesDirectory}
	return &Resolver{
		Loader:              uriLoader,
		DefaultLoader:       defaultLoader,
		Registry:            opts.Registry,
		References:          opts.References,
		FallbackLanguageTag: opts.FallbackLanguageTag,
	}
}

func (r *Resolver) Resolve(ctx *ResolveContext, typ string) (resolved *Resolved, err error) {
	t, ok := r.Registry.Lookup(typ)
	if !ok {
		panic("template: unregistered template type: " + typ)
	}

	content, err := r.loadContent(ctx, t)
	if err != nil {
		return
	}

	var translations map[string]Translation
	if t.TranslationTemplateType != "" {
		translations, err = r.ResolveTranslations(ctx, t.TranslationTemplateType)
		if err != nil {
			return
		}
	}

	componentContents, err := r.resolveComponents(ctx, t.ComponentTemplateTypes)
	if err != nil {
		return
	}

	resolved = &Resolved{
		T:                 t,
		Content:           content,
		Translations:      translations,
		ComponentContents: componentContents,
	}
	return
}

func (r *Resolver) loadContent(ctx *ResolveContext, t T) (string, error) {
	ref, err := r.resolveReference(ctx, t)
	if err == nil {
		return r.Loader.Load(ref.URI)
	}

	return r.DefaultLoader.LoadDefault(t.Type)
}

type referenceLanguageTagger Reference

func (r referenceLanguageTagger) GetLanguageTag() string {
	return r.LanguageTag
}

func (r *Resolver) resolveReference(ctx *ResolveContext, t T) (reference *Reference, err error) {
	var refs []Reference

	// Find out references with the target type.
	for _, ref := range r.References {
		if ref.Type == t.Type {
			rr := ref
			refs = append(refs, rr)
		}
	}

	// We either have a list of references of different language tags or an empty list.
	if len(refs) <= 0 {
		err = &errNotFound{name: t.Type}
		return
	}

	var items []languageTagger
	for _, ref := range refs {
		if ref.LanguageTag == "" {
			ref.LanguageTag = string(intl.Fallback(r.FallbackLanguageTag))
		}
		items = append(items, referenceLanguageTagger(ref))
	}

	matched, err := languageMatch(ctx.PreferredLanguageTags, r.FallbackLanguageTag, items)
	if err != nil {
		return
	}

	tagger := (*matched).(referenceLanguageTagger)
	ref := Reference(tagger)
	reference = &ref
	return
}

func (r *Resolver) ResolveTranslations(ctx *ResolveContext, typ string) (translations map[string]Translation, err error) {
	t, ok := r.Registry.Lookup(typ)
	if !ok {
		panic("template: unregistered template type: " + typ)
	}

	keyToTagToTranslation := make(map[string]map[string]string)
	insert := func(tag string, translation map[string]string) {
		for key, val := range translation {
			m, ok := keyToTagToTranslation[key]
			if !ok {
				keyToTagToTranslation[key] = make(map[string]string)
				m = keyToTagToTranslation[key]
			}
			m[tag] = val
		}
	}

	// Load the default translation
	defaultJSON, err := r.DefaultLoader.LoadDefault(typ)
	if err != nil {
		return
	}
	defaultTranslation, err := loadTranslation(defaultJSON)
	if err != nil {
		return
	}
	insert(intl.DefaultLanguage, defaultTranslation)

	// Find out all references.
	var refs []Reference
	for _, ref := range r.References {
		if ref.Type == t.Type {
			rr := ref
			refs = append(refs, rr)
		}
	}

	// Load all provided translations
	for _, ref := range refs {
		var jsonStr string
		jsonStr, err = r.Loader.Load(ref.URI)
		if err != nil {
			return
		}
		var translation map[string]string
		translation, err = loadTranslation(jsonStr)
		if err != nil {
			return
		}

		tag := ref.LanguageTag
		if tag == "" {
			tag = string(intl.Fallback(r.FallbackLanguageTag))
		}
		insert(tag, translation)
	}

	// Finally, we resolve each key.
	translations = make(map[string]Translation)
	for key, availableTransltion := range keyToTagToTranslation {
		var items []languageTagger
		for languageTag, value := range availableTransltion {
			items = append(items, Translation{
				LanguageTag: languageTag,
				Value:       value,
			})
		}
		var matched *languageTagger
		matched, err = languageMatch(ctx.PreferredLanguageTags, r.FallbackLanguageTag, items)
		if err != nil {
			return
		}

		tagger := (*matched).(Translation)
		translations[key] = tagger
	}

	return
}

func (r *Resolver) resolveComponents(ctx *ResolveContext, types []string) (contents []string, err error) {
	resolvedContents := make(map[string]string)

	// We need to declare it first otherwise recur cannot reference itself.
	var recur func(types []string) (err error)

	recur = func(types []string) (err error) {
		for _, typ := range types {
			// Do not need to load the same type more than once.
			_, ok := resolvedContents[typ]
			if ok {
				continue
			}

			t, ok := r.Registry.Lookup(typ)
			if !ok {
				panic("template: unregistered template type: " + typ)
			}

			var content string
			content, err = r.loadContent(ctx, t)
			if err != nil {
				return
			}

			resolvedContents[typ] = content

			err = recur(t.ComponentTemplateTypes)
			if err != nil {
				return
			}
		}
		return
	}

	err = recur(types)
	if err != nil {
		return
	}

	for _, typ := range types {
		contents = append(contents, resolvedContents[typ])
	}
	return
}

func loadTranslation(jsonStr string) (translation map[string]string, err error) {
	var jsonObj map[string]interface{}
	err = json.Unmarshal([]byte(jsonStr), &jsonObj)
	if err != nil {
		err = fmt.Errorf("template: expected translation file to be JSON: %w", err)
		return
	}

	translation = map[string]string{}
	for key, val := range jsonObj {
		s, ok := val.(string)
		if !ok {
			err = fmt.Errorf("template: expected translation value to be string: %s %T", key, val)
			return
		}
		translation[key] = s
	}
	return
}
