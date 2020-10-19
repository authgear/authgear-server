package template

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/authgear/authgear-server/pkg/util/resource"
)

type Translation struct {
	LanguageTag string
	Value       string
}

func (t Translation) GetLanguageTag() string {
	return t.LanguageTag
}

const TranslationJSONName = "translation.json"

type translationJSON struct{}

var TranslationJSON = resource.RegisterResource(&translationJSON{})

func (t *translationJSON) ReadResource(fs resource.Fs) ([]resource.LayerFile, error) {
	return readTemplates(fs, TranslationJSONName)
}

func (t *translationJSON) MatchResource(path string) bool {
	return matchTemplatePath(path, TranslationJSONName)
}

func (t *translationJSON) Merge(layers []resource.LayerFile, args map[string]interface{}) (*resource.MergedFile, error) {
	return mergeTranslations(layers, args)
}

func (t *translationJSON) Parse(merged *resource.MergedFile) (interface{}, error) {
	var translations map[string]Translation
	if err := json.Unmarshal(merged.Data, &translations); err != nil {
		return nil, err
	}
	return translations, nil
}

func mergeTranslations(layers []resource.LayerFile, args map[string]interface{}) (*resource.MergedFile, error) {
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

	translationData := make(map[string]Translation)
	for key, translations := range translationMap {
		if _, ok := translations[LanguageTag(defaultLanguageTag)]; !ok {
			translations[LanguageTag(defaultLanguageTag)] = translations[LanguageTagDefault]
		}
		delete(translations, LanguageTagDefault)

		var items []LanguageItem
		for languageTag, value := range translations {
			items = append(items, Translation{
				LanguageTag: string(languageTag),
				Value:       string(value),
			})
		}
		var matched LanguageItem
		matched, err := MatchLanguage(preferredLanguageTags, defaultLanguageTag, items)
		if errors.Is(err, ErrNoLanguageMatch) {
			if len(items) > 0 {
				// Use first item in case of no match, to ensure resolution always succeed
				matched = items[0]
			} else {
				// Ignore keys without translation
				continue
			}
		} else if err != nil {
			return nil, err
		}

		translationData[string(key)] = matched.(Translation)
	}

	var mergedData interface{}
	if mergeRaw, _ := args[resource.ArgMergeRaw].(bool); mergeRaw {
		rawData := make(map[string]string)
		for key, t := range translationData {
			rawData[key] = t.Value
		}
		mergedData = rawData
	} else {
		mergedData = translationData
	}

	data, err := json.Marshal(mergedData)
	if err != nil {
		return nil, err
	}

	return &resource.MergedFile{Data: data}, nil
}
