package template

import (
	"sort"

	"golang.org/x/text/language"

	"github.com/skygeario/skygear-server/pkg/core/config"
)

type RenderOptions struct {
	Required bool
	Key      string
}

type NewEngineOptions struct {
	EnableFileLoader      bool
	EnableDataLoader      bool
	AssetGearLoader       *AssetGearLoader
	TemplateItems         []config.TemplateItem
	PreferredLanguageTags []string
}

// Engine resolves and renders templates.
type Engine struct {
	uriLoader     *URILoader
	TemplateSpecs map[config.TemplateItemType]Spec
	templateItems []config.TemplateItem
	// NOTE(louis): PreferredLanguageTags
	// preferredLanguageTags is put here instead of receiving it from RenderXXX methods.
	// It is expected that engine is created per request.
	// The preferred language tags should be lazily retrieved
	// from the auth context.
	preferredLanguageTags []string
}

func NewEngine(opts NewEngineOptions) *Engine {
	uriLoader := NewURILoader(opts.AssetGearLoader)
	uriLoader.EnableFileLoader = opts.EnableFileLoader
	uriLoader.EnableDataLoader = opts.EnableDataLoader
	return &Engine{
		uriLoader:             uriLoader,
		templateItems:         opts.TemplateItems,
		TemplateSpecs:         map[config.TemplateItemType]Spec{},
		preferredLanguageTags: opts.PreferredLanguageTags,
	}
}

func (e *Engine) Register(spec Spec) {
	e.TemplateSpecs[spec.Type] = spec
}

func (e *Engine) RenderTemplate(templateType config.TemplateItemType, context map[string]interface{}, option RenderOptions) (out string, err error) {
	templateBody, spec, err := e.resolveTemplate(templateType, option)
	if err != nil {
		return
	}
	if spec.IsHTML {
		return RenderHTMLTemplate(string(templateType), templateBody, context)
	}
	return RenderTextTemplate(string(templateType), templateBody, context)
}

func (e *Engine) resolveTemplate(templateType config.TemplateItemType, options RenderOptions) (string, Spec, error) {
	spec, found := e.TemplateSpecs[templateType]
	if !found {
		panic("template: unregistered template type: " + templateType)
	}

	templateItem, err := e.resolveTemplateItem(spec, options.Key)
	var templateBody string
	// No template item can be resolved. Fallback to default.
	if err != nil {
		err = nil
		if spec.Default != "" {
			templateBody = spec.Default
		} else if options.Required {
			err = &errNotFound{string(templateType)}
		}
	} else {
		templateBody, err = e.uriLoader.Load(templateItem.URI)
	}
	return templateBody, spec, err
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

	// We have a list of templates of different language tags.
	// The first item in tags is used as fallback.
	// So we have sort the templates so that template with empty
	// language tag comes first.
	//
	// language.Make("") is "und"
	sort.Slice(input, func(i, j int) bool {
		return input[i].LanguageTag < input[j].LanguageTag
	})

	supportedTags := make([]language.Tag, len(input))
	for i, item := range input {
		supportedTags[i] = language.Make(item.LanguageTag)
	}
	matcher := language.NewMatcher(supportedTags)

	preferredTags := make([]language.Tag, len(e.preferredLanguageTags))
	for i, tagStr := range e.preferredLanguageTags {
		preferredTags[i] = language.Make(tagStr)
	}

	_, idx, _ := matcher.Match(preferredTags...)

	return &input[idx], nil
}
