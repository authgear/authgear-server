package template

import (
	"sort"

	"golang.org/x/text/language"

	"github.com/skygeario/skygear-server/pkg/core/config"
)

type ResolveOptions struct {
	Key string
}

type NewEngineOptions struct {
	EnableFileLoader bool
	EnableDataLoader bool
	AssetGearLoader  *AssetGearLoader
	TemplateItems    []config.TemplateItem
}

// Engine resolves and renders templates.
type Engine struct {
	uriLoader             *URILoader
	TemplateSpecs         map[config.TemplateItemType]Spec
	templateItems         []config.TemplateItem
	preferredLanguageTags []string
}

func NewEngine(opts NewEngineOptions) *Engine {
	uriLoader := NewURILoader(opts.AssetGearLoader)
	uriLoader.EnableFileLoader = opts.EnableFileLoader
	uriLoader.EnableDataLoader = opts.EnableDataLoader
	return &Engine{
		uriLoader:     uriLoader,
		templateItems: opts.TemplateItems,
		TemplateSpecs: map[config.TemplateItemType]Spec{},
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

// Register registers spec with e.
func (e *Engine) Register(spec Spec) {
	e.TemplateSpecs[spec.Type] = spec
}

func (e *Engine) RenderTemplate(templateType config.TemplateItemType, context map[string]interface{}, resolveOptions ResolveOptions, validateOpts ...ValidatorOption) (out string, err error) {
	templateBody, spec, err := e.resolveTemplate(templateType, resolveOptions)
	if err != nil {
		return
	}
	if spec.IsHTML {
		return RenderHTMLTemplate(RenderOptions{
			Name:          string(templateType),
			TemplateBody:  templateBody,
			Defines:       spec.Defines,
			Context:       context,
			ValidatorOpts: validateOpts,
		})
	}
	return RenderTextTemplate(RenderOptions{
		Name:          string(templateType),
		TemplateBody:  templateBody,
		Defines:       spec.Defines,
		Context:       context,
		ValidatorOpts: validateOpts,
	})
}

func (e *Engine) resolveTemplate(templateType config.TemplateItemType, options ResolveOptions) (templateBody string, spec Spec, err error) {
	spec, ok := e.TemplateSpecs[templateType]
	if !ok {
		panic("template: unregistered template type: " + templateType)
	}

	templateBody = spec.Default
	templateItem, err := e.resolveTemplateItem(spec, options.Key)
	// No template item can be resolved. Fallback to default.
	if err != nil {
		err = nil
		return
	}

	templateBody, err = e.uriLoader.Load(templateItem.URI)
	if err != nil {
		return
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
