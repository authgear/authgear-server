package template

import (
	"encoding/json"
	"fmt"

	"github.com/authgear/authgear-server/pkg/util/resource"
	"github.com/authgear/authgear-server/pkg/util/template"
)

const translationJSONName = "translation.json"

type translationJSON struct{}

var TranslationJSON = resource.RegisterResource(&translationJSON{})

func (t *translationJSON) ReadResource(fs resource.Fs) ([]resource.LayerFile, error) {
	return readTemplates(fs, translationJSONName)
}

func (t *translationJSON) MatchResource(path string) bool {
	return matchTemplatePath(path, translationJSONName)
}

func (t *translationJSON) Merge(layers []resource.LayerFile, args map[string]interface{}) (*resource.LayerFile, error) {
	return mergeTranslations(layers, args)
}

func (t *translationJSON) Parse(data []byte) (interface{}, error) {
	var translations map[string]template.Translation
	if err := json.Unmarshal(data, &translations); err != nil {
		return nil, err
	}
	return translations, nil
}

func mergeTranslations(layers []resource.LayerFile, args map[string]interface{}) (*resource.LayerFile, error) {
	preferredLanguageTags, _ := args[ResourceArgPreferredLanguageTag].([]string)
	defaultLanguageTag, _ := args[ResourceArgDefaultLanguageTag].(string)

	type LanguageTag string
	type TranslationKey string
	type TranslationValue string

	translationMap := make(map[TranslationKey]map[LanguageTag]TranslationValue)
	insertTranslation := func(tag, key, value string) {
		keyTranslations, ok := translationMap[TranslationKey(key)]
		if !ok {
			keyTranslations = make(map[LanguageTag]TranslationValue)
			translationMap[TranslationKey(key)] = keyTranslations
		}
		keyTranslations[LanguageTag(tag)] = TranslationValue(value)
	}

	for _, file := range layers {
		langTag := templateLanguageTagRegex.FindStringSubmatch(file.Path)[1]

		var jsonObj map[string]interface{}
		if err := json.Unmarshal(file.Data, &jsonObj); err != nil {
			return nil, fmt.Errorf("translation file must be JSON: %w", err)
		}

		for key, val := range jsonObj {
			value, ok := val.(string)
			if !ok {
				return nil, fmt.Errorf("translation value must be string: %s %T", key, val)
			}
			insertTranslation(langTag, key, value)
		}
	}

	translationData := make(map[string]template.Translation)
	for key, translations := range translationMap {
		if _, ok := translations[LanguageTag(defaultLanguageTag)]; !ok {
			translations[LanguageTag(defaultLanguageTag)] = translations[LanguageTagDefault]
		}
		delete(translations, LanguageTagDefault)

		var items []template.LanguageItem
		for languageTag, value := range translations {
			items = append(items, template.Translation{
				LanguageTag: string(languageTag),
				Value:       string(value),
			})
		}
		var matched template.LanguageItem
		matched, err := template.MatchLanguage(preferredLanguageTags, defaultLanguageTag, items)
		if err != nil {
			return nil, err
		}

		translationData[string(key)] = matched.(template.Translation)
	}

	data, err := json.Marshal(translationData)
	if err != nil {
		return nil, err
	}

	return &resource.LayerFile{
		Path: "templates/" + translationJSONName,
		Data: data,
	}, nil
}
