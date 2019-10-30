package template

import (
	"github.com/skygeario/skygear-server/pkg/core/config"
)

// Engine resolves and renders templates.
type Engine struct {
	DefaultLoader *StringLoader
	URILoader     *URILoader
}

type RenderOptions struct {
	Required bool
	Key      string
}

func NewEngine(enabledFileLoader bool) *Engine {
	return &Engine{
		DefaultLoader: NewStringLoader(),
		URILoader:     NewURILoader(enabledFileLoader),
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

func (e *Engine) resolveTemplate(templateType config.TemplateItemType, option RenderOptions) (templateBody string, err error) {
	// TODO(template)
	panic("TODO")
}
