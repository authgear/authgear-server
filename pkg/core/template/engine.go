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

func (e *Engine) RegisterDefaultTemplate(templateName string, template string) {
	e.defaultLoader.StringMap[templateName] = template
}

func (e *Engine) ParseTextTemplate(templateName string, context map[string]interface{}, required bool) (out string, err error) {
	var templateBody string
	if templateBody, err = e.downloadContent(templateName, required); err != nil {
		return
	}

	return ParseTextTemplate(templateBody, context)
}

func (e *Engine) ParseHTMLTemplate(templateName string, context map[string]interface{}, required bool) (out string, err error) {
	var templateBody string
	if templateBody, err = e.downloadContent(templateName, required); err != nil {
		return
	}

	return ParseHTMLTemplate(templateBody, context)
}

func (e *Engine) allLoaders() []Loader {
	return append(e.loaders, e.defaultLoader)
}

func (e *Engine) downloadContent(templateName string, required bool) (templateBody string, err error) {
	loaders := e.allLoaders()
	for _, loader := range loaders {
		if templateBody, err = loader.Load(templateName); err == nil {
			break
		}
	}

	if required && err != nil {
		// return error if required but template not found
		err = fmt.Errorf("template with name `%s` not found", templateName)
	} else if !required && err != nil {
		// no error if not required
		err = nil
		templateBody = ""
	}

	return
}
