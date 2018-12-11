package template

import (
	"fmt"
)

// Engine parse templates with given url, and fallback to a default one if the
// given one not found
type Engine struct {
	defaultPathMap map[string]string
	urlMap         map[string]string
}

func NewEngine() *Engine {
	return &Engine{
		defaultPathMap: make(map[string]string),
		urlMap:         make(map[string]string),
	}
}

func NewEngineFromEngine(engine *Engine) *Engine {
	newEngine := NewEngine()
	for k, v := range engine.defaultPathMap {
		newEngine.defaultPathMap[k] = v
	}

	for k, v := range engine.urlMap {
		newEngine.urlMap[k] = v
	}

	return newEngine
}

func (e *Engine) RegisterDefaultTemplate(templateName string, defaultPath string) {
	e.defaultPathMap[templateName] = defaultPath
}

func (e *Engine) RegisterTemplate(templateName string, url string) {
	e.urlMap[templateName] = url
}

func (e *Engine) ParseTextTemplate(templateName string, context map[string]interface{}, required bool) (out string, err error) {
	return e.parseWithDefaultFallback(ParseTextTemplate, templateName, context, required)
}

func (e *Engine) ParseHTMLTemplate(templateName string, context map[string]interface{}, required bool) (out string, err error) {
	return e.parseWithDefaultFallback(ParseHTMLTemplate, templateName, context, required)
}

func (e *Engine) parseWithDefaultFallback(
	parseFunc func(string, map[string]interface{}) (string, error),
	templateName string,
	context map[string]interface{},
	required bool,
) (out string, err error) {
	var templateBody string
	url, found := e.urlMap[templateName]
	if found {
		if templateBody, err = DownloadTemplateFromURL(url); err == nil {
			out, err = parseFunc(templateBody, context)
			return
		}
	}

	url, found = e.defaultPathMap[templateName]
	if !found {
		// if require = false, no error would be thrown for template not found
		if !required {
			out, err = "", nil
			return
		}

		panic(fmt.Errorf("unexpected default template with name `%s` not found", templateName))
	}

	if templateBody, err = DownloadTemplateFromFilePath(url); err != nil {
		panic(fmt.Errorf("unexpected unable to get content of default template with name `%s`", templateName))
	}

	out, err = parseFunc(templateBody, context)
	return
}
