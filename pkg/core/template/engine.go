package template

import (
	"fmt"
)

// Engine parse templates with given url, and fallback to a default one if the
// given one not found
//
// load order follow the same order of loaders
type Engine struct {
	defaultLoader *StringLoader
	loaders       []Loader
}

type ParseOption struct {
	Required             bool
	FallbackTemplateName string
}

func NewEngine() *Engine {
	return &Engine{
		defaultLoader: NewStringLoader(),
		loaders:       []Loader{},
	}
}

func (e *Engine) CopyDefaultToEngine(engine *Engine) {
	for k, v := range e.defaultLoader.StringMap {
		engine.defaultLoader.StringMap[k] = v
	}
}

func (e *Engine) SetLoaders(loaders []Loader) {
	e.loaders = loaders
}

func (e *Engine) PrependLoader(loader Loader) {
	e.loaders = append([]Loader{loader}, e.loaders...)
}

func (e *Engine) RegisterDefaultTemplate(templateName string, template string) {
	e.defaultLoader.StringMap[templateName] = template
}

func (e *Engine) ParseTextTemplate(templateName string, context map[string]interface{}, option ParseOption) (out string, err error) {
	var templateBody string
	if templateBody, err = e.downloadContent(templateName, option); err != nil {
		return
	}

	return ParseTextTemplate(templateBody, context)
}

func (e *Engine) ParseHTMLTemplate(templateName string, context map[string]interface{}, option ParseOption) (out string, err error) {
	var templateBody string
	if templateBody, err = e.downloadContent(templateName, option); err != nil {
		return
	}

	return ParseHTMLTemplate(templateBody, context)
}

func (e *Engine) downloadContent(templateName string, option ParseOption) (templateBody string, err error) {
	// skip error handling if there is a fallback template name
	if option.FallbackTemplateName == "" {
		defer func() {
			if option.Required && err != nil {
				// return error if required but template not found
				err = fmt.Errorf("template with name `%s` not found", templateName)
			} else if !option.Required && err != nil {
				// no error if not required
				err = nil
				templateBody = ""
			}
		}()
	}

	for _, loader := range e.loaders {
		if templateBody, err = loader.Load(templateName); err == nil {
			return
		}
	}

	// try with default loader
	templateBody, err = e.defaultLoader.Load(templateName)

	if option.FallbackTemplateName != "" && err != nil {
		// download with fallback template name
		templateBody, err = e.downloadContent(option.FallbackTemplateName, ParseOption{
			Required:             option.Required,
			FallbackTemplateName: "",
		})
	}

	return
}
