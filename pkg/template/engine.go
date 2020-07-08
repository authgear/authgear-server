package template

import (
	"encoding/json"
	"fmt"

	"github.com/authgear/authgear-server/pkg/auth/config"
	"github.com/authgear/authgear-server/pkg/core/intl"
	"github.com/iawaknahc/gomessageformat"
)

type ResolveOptions struct {
	Key string
}

type resolveResult struct {
	Spec         Spec
	TemplateBody string
	// Translations is key -> tag -> translation.
	// For example,
	// {
	//   "key1": {
	//     "en": "Hello",
	//     "en-US": "Hi!",
	//     "zh": "你好"
	//   }
	// }
	Translations map[string]map[string]string
	Components   []string
}

type NewEngineOptions struct {
	TemplateItems    []config.TemplateItem
	FallbackLanguage string
}

type Loader interface {
	Load(string) (string, error)
}

// Engine resolves and renders templates.
type Engine struct {
	loader                Loader
	TemplateSpecs         map[config.TemplateItemType]Spec
	templateItems         []config.TemplateItem
	preferredLanguageTags []string
	fallbackLanguageTag   string
	validatorOptions      []ValidatorOption
}

func NewEngine(opts NewEngineOptions) *Engine {
	uriLoader := NewURILoader()
	return &Engine{
		loader:              uriLoader,
		templateItems:       opts.TemplateItems,
		TemplateSpecs:       map[config.TemplateItemType]Spec{},
		fallbackLanguageTag: opts.FallbackLanguage,
	}
}

// Clone clones e.
func (e *Engine) Clone() *Engine {
	// A simply struct copy is enough here because we assume
	// Register calls are made only during engine creation.
	newEngine := *e
	return &newEngine
}

// WithPreferredLanguageTags returns a new engine with the given tags.
// This function offers greater flexibility on configuring preferred languages because
// This information may not be available at the creation of the engine.
func (e *Engine) WithPreferredLanguageTags(tags []string) *Engine {
	newEngine := e.Clone()
	newEngine.preferredLanguageTags = tags
	return newEngine
}

// WithValidatorOptions returns a new engine with the givan validator options.
func (e *Engine) WithValidatorOptions(opts ...ValidatorOption) *Engine {
	newEngine := e.Clone()
	newEngine.validatorOptions = opts
	return newEngine
}

// Register registers spec with e.
func (e *Engine) Register(spec Spec) {
	e.TemplateSpecs[spec.Type] = spec
}

func (e *Engine) RenderTemplate(templateType config.TemplateItemType, data interface{}, resolveOptions ResolveOptions) (out string, err error) {
	result, err := e.resolveTemplate(templateType, resolveOptions)
	if err != nil {
		return
	}

	defines := []string{}
	defines = append(defines, result.Spec.Defines...)
	defines = append(defines, result.Components...)
	renderOptions := RenderOptions{
		Name:          string(templateType),
		TemplateBody:  result.TemplateBody,
		Defines:       defines,
		Data:          data,
		ValidatorOpts: e.validatorOptions,
	}

	if result.Spec.Translation != "" {
		renderOptions.Funcs = map[string]interface{}{
			"localize": makeLocalize(
				e.preferredLanguageTags,
				e.fallbackLanguageTag,
				result.Translations,
			),
		}
	}

	renderFunc := RenderTextTemplate
	if result.Spec.IsHTML {
		renderFunc = RenderHTMLTemplate
	}

	return renderFunc(renderOptions)
}

func (e *Engine) resolveTemplate(templateType config.TemplateItemType, options ResolveOptions) (result *resolveResult, err error) {
	spec, ok := e.TemplateSpecs[templateType]
	if !ok {
		panic("template: unregistered template type: " + templateType)
	}

	templateBody, err := e.loadTemplateBody(spec, options.Key)
	if err != nil {
		return
	}

	// Resolve the translations, if any
	var translations map[string]map[string]string
	if spec.Translation != "" {
		translations, err = e.resolveTranslations(spec.Translation)
		if err != nil {
			return
		}
	}

	// Resolve components
	components, err := e.resolveComponents(spec.Components, options.Key)
	if err != nil {
		return
	}

	result = &resolveResult{
		Spec:         spec,
		TemplateBody: templateBody,
		Translations: translations,
		Components:   components,
	}

	return
}

func (e *Engine) loadTemplateBody(spec Spec, key string) (templateBody string, err error) {
	// Take the default value by default
	templateBody = spec.Default
	templateItem, err := e.resolveTemplateItem(spec, key)
	if err != nil {
		// No template item can be resolved. Fallback to default.
		err = nil
	} else {
		templateBody, err = e.loader.Load(templateItem.URI)
		if err != nil {
			return
		}
	}
	return
}

func (e *Engine) resolveTemplateItem(spec Spec, key string) (templateItem *config.TemplateItem, err error) {
	input := e.templateItems
	var output []config.TemplateItem

	// The first step is to find out templates with the target type.
	for _, item := range input {
		if item.Type == spec.Type {
			i := item
			output = append(output, i)
		}
	}
	input = output
	output = nil

	// The second step is to find out templates with the target key, if key is specified
	if spec.IsKeyed && key != "" {
		for _, item := range input {
			if item.Key == key {
				i := item
				output = append(output, i)
			}
		}
		input = output
	}

	// We have either have a list of templates of different language tags or an empty list.
	if len(input) <= 0 {
		err = &errNotFound{name: string(spec.Type)}
		return
	}

	// Set the empty language tag to fallback language.
	for i, item := range input {
		if item.LanguageTag == "" {
			item.LanguageTag = string(intl.Fallback(e.fallbackLanguageTag))
			input[i] = item
		}
	}

	// Build a map from language tag to item and supported languages tags.
	languageTagToItem := make(map[string]config.TemplateItem)
	var rawSupported []string
	for _, item := range input {
		languageTagToItem[item.LanguageTag] = item
		rawSupported = append(rawSupported, item.LanguageTag)
	}
	supportedLanguageTags := intl.Supported(rawSupported, intl.Fallback(e.fallbackLanguageTag))

	idx, _ := intl.Match(e.preferredLanguageTags, supportedLanguageTags)
	tag := supportedLanguageTags[idx]

	item, ok := languageTagToItem[tag]
	if !ok {
		err = &errNotFound{name: string(spec.Type)}
		return
	}

	return &item, nil
}

func (e *Engine) resolveTranslations(templateType config.TemplateItemType) (translations map[string]map[string]string, err error) {
	spec, ok := e.TemplateSpecs[templateType]
	if !ok {
		panic("template: unregistered template type: " + templateType)
	}

	translations = map[string]map[string]string{}

	// Load the default translation
	defaultTranslation, err := loadTranslation(spec.Default)
	if err != nil {
		return
	}
	insertTranslation(translations, intl.DefaultLanguage, defaultTranslation)

	// Find out all items
	var items []config.TemplateItem
	for _, item := range e.templateItems {
		if item.Type == spec.Type {
			i := item
			items = append(items, i)
		}
	}

	// Load all provided translations
	for _, item := range items {
		var jsonStr string
		jsonStr, err = e.loader.Load(item.URI)
		if err != nil {
			return
		}
		var translation map[string]string
		translation, err = loadTranslation(jsonStr)
		if err != nil {
			return
		}
		insertTranslation(translations, item.LanguageTag, translation)
	}

	return
}

func (e *Engine) resolveComponents(types []config.TemplateItemType, key string) (bodies []string, err error) {
	resolvedBodies := make(map[config.TemplateItemType]string)

	// We need to declare it first otherwise recur cannot reference itself.
	var recur func(types []config.TemplateItemType) (err error)

	recur = func(types []config.TemplateItemType) (err error) {
		for _, templateType := range types {
			// Do not need to load the same type more than once.
			_, ok := resolvedBodies[templateType]
			if ok {
				continue
			}

			spec, ok := e.TemplateSpecs[templateType]
			if !ok {
				panic("template: unregistered template type: " + templateType)
			}
			var body string
			body, err = e.loadTemplateBody(spec, key)
			if err != nil {
				return
			}

			resolvedBodies[templateType] = body

			err = recur(spec.Components)
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

	for _, body := range resolvedBodies {
		bodies = append(bodies, body)
	}
	return
}

func makeLocalize(preferredLanguageTags []string, fallbackLanguageTag string, translations map[string]map[string]string) func(key string, args ...interface{}) (string, error) {
	return func(key string, args ...interface{}) (out string, err error) {
		m, ok := translations[key]
		if !ok {
			err = fmt.Errorf("translation key not found: %s", key)
			return
		}

		var supportedLanguageTags []string
		for tag := range m {
			supportedLanguageTags = append(supportedLanguageTags, tag)
		}
		supportedLanguageTags = intl.Supported(supportedLanguageTags, intl.Fallback(fallbackLanguageTag))

		idx, tag := intl.Match(preferredLanguageTags, supportedLanguageTags)
		pattern := m[supportedLanguageTags[idx]]

		out, err = messageformat.FormatPositional(tag, pattern, args...)
		if err != nil {
			return
		}

		return
	}
}

func loadTranslation(jsonStr string) (translation map[string]string, err error) {
	var jsonObj map[string]interface{}
	err = json.Unmarshal([]byte(jsonStr), &jsonObj)
	if err != nil {
		err = fmt.Errorf("expected translation file to be JSON: %w", err)
		return
	}

	translation = map[string]string{}
	for key, val := range jsonObj {
		s, ok := val.(string)
		if !ok {
			err = fmt.Errorf("expected translation value to be string: %s %T", key, val)
			return
		}
		translation[key] = s
	}
	return
}

func insertTranslation(translations map[string]map[string]string, tag string, translation map[string]string) {
	for key, val := range translation {
		m, ok := translations[key]
		if !ok {
			translations[key] = map[string]string{}
			m = translations[key]
		}
		m[tag] = val
	}
}
