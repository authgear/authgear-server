package main

import (
	"encoding/json"
	htmltemplate "html/template"
	"os"
	"path/filepath"
	"slices"
	texttemplate "text/template"

	. "github.com/authgear/authgear-server/pkg/util/template"
)

type EngineTemplateResolverImpl struct{}

func (r *EngineTemplateResolverImpl) ResolveHTML(desc *HTML, preferredLanguages []string) (tpl *htmltemplate.Template, err error) {
	return
}

func (r *EngineTemplateResolverImpl) ResolvePlainText(desc *PlainText, preferredLanguages []string) (tpl *texttemplate.Template, err error) {
	return
}

func (r *EngineTemplateResolverImpl) ResolveTranslations(preferredLanguages []string) (translations map[string]Translation, err error) {
	translations = make(map[string]Translation)

	matches, err := filepath.Glob("./resources/authgear/templates/*/translation.json")
	if err != nil {
		return
	}

	for _, match := range matches {
		langTag := filepath.Base(filepath.Dir(match))
		if !slices.Contains(preferredLanguages, langTag) {
			continue
		}

		var f *os.File
		f, err = os.Open(match)
		if err != nil {
			return
		}
		defer f.Close()

		var jsonObject map[string]interface{}
		err = json.NewDecoder(f).Decode(&jsonObject)
		if err != nil {
			return
		}

		for key, value := range jsonObject {
			translations[key] = Translation{
				LanguageTag: langTag,
				Value:       value.(string),
			}
		}
	}

	return
}
