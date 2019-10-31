package template

import (
	"sort"

	"golang.org/x/text/language"

	"github.com/skygeario/skygear-server/pkg/core/config"
)

// Engine resolves and renders templates.
type Engine struct {
	DefaultLoader         *StringLoader
	URILoader             *URILoader
	TemplateItems         []config.TemplateItem
	PreferredLanguageTags []string
}

type RenderOptions struct {
	Required bool
	Key      string
}

func NewEngine(
	fileLoaderEnabled bool,
	dataLoaderEnabled bool,
	templateItems []config.TemplateItem,
	tags []string,
) *Engine {
	return &Engine{
		DefaultLoader:         NewStringLoader(),
		URILoader:             NewURILoader(fileLoaderEnabled, dataLoaderEnabled),
		TemplateItems:         templateItems,
		PreferredLanguageTags: tags,
	}
}

func (e *Engine) Clone() *Engine {
	items := make([]config.TemplateItem, len(e.TemplateItems))
	tags := make([]string, len(e.PreferredLanguageTags))
	copy(items, e.TemplateItems)
	copy(tags, e.PreferredLanguageTags)

	return &Engine{
		DefaultLoader:         e.DefaultLoader.Clone(),
		URILoader:             e.URILoader.Clone(),
		TemplateItems:         items,
		PreferredLanguageTags: tags,
	}
}

func (e *Engine) SetDefault(templateType config.TemplateItemType, template string) {
	e.DefaultLoader.StringMap[string(templateType)] = template
}

func (e *Engine) RenderTextTemplate(templateType config.TemplateItemType, context map[string]interface{}, option RenderOptions) (out string, err error) {
	var templateBody string
	if templateBody, err = e.resolveTemplate(templateType, option); err != nil {
		return
	}

	return RenderTextTemplate(string(templateType), templateBody, context)
}

func (e *Engine) RenderHTMLTemplate(templateType config.TemplateItemType, context map[string]interface{}, option RenderOptions) (out string, err error) {
	var templateBody string
	if templateBody, err = e.resolveTemplate(templateType, option); err != nil {
		return
	}

	return RenderHTMLTemplate(string(templateType), templateBody, context)
}

func (e *Engine) resolveTemplate(templateType config.TemplateItemType, options RenderOptions) (string, error) {
	templateItem, err := e.resolveTemplateItem(templateType, options.Key)
	// No template item can be resolved. Fallback to default.
	if err != nil {
		templateBody, err := e.DefaultLoader.Load(string(templateType))
		if err != nil {
			if !options.Required {
				err = nil
			}
		}
		return templateBody, err
	}
	return e.URILoader.Load(templateItem.URI)
}

func (e *Engine) resolveTemplateItem(templateType config.TemplateItemType, key string) (templateItem *config.TemplateItem, err error) {
	input := e.TemplateItems
	var output []config.TemplateItem

	// The first step is to find out templates with the target type.
	for _, item := range input {
		if item.Type == templateType {
			i := item
			output = append(output, i)
		}
	}
	input = output
	output = nil

	// The second step is to find out templates with the target key, if key is specified
	if key != "" {
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
		err = &errNotFound{name: string(templateType)}
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

	preferredTags := make([]language.Tag, len(e.PreferredLanguageTags))
	for i, tagStr := range e.PreferredLanguageTags {
		preferredTags[i] = language.Make(tagStr)
	}

	_, idx, _ := matcher.Match(preferredTags...)

	return &input[idx], nil
}
