package template

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"

	messageformat "github.com/iawaknahc/gomessageformat"
	"golang.org/x/text/language"

	"github.com/authgear/authgear-server/pkg/util/intlresource"
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

var appSpecificKeysRegex = []*regexp.Regexp{
	regexp.MustCompile(`^app\.name$`),
	regexp.MustCompile(`^email\..+\.sender$`),
	regexp.MustCompile(`^email\..+\.reply-to$`),
	regexp.MustCompile(`^sms\..+\.sender$`),
}

var fsLevelsOrderedInAscendingPriority = []resource.FsLevel{
	resource.FsLevelBuiltin,
	resource.FsLevelCustom,
	resource.FsLevelApp,
}

type translationJSON struct{}

var _ resource.Descriptor = &translationJSON{}

var TranslationJSON = resource.RegisterResource(&translationJSON{})

func (t *translationJSON) IsAppSpecificKey(key string) bool {
	for _, r := range appSpecificKeysRegex {
		matched := r.MatchString(key)
		if matched {
			return true
		}
	}
	return false
}

func (t *translationJSON) MatchResource(path string) (*resource.Match, bool) {
	return matchTemplatePath(path, TranslationJSONName)
}

func (t *translationJSON) FindResources(fs resource.Fs) ([]resource.Location, error) {
	return readTemplates(fs, TranslationJSONName)
}

func (t *translationJSON) ViewResources(resources []resource.ResourceFile, rawView resource.View) (interface{}, error) {
	switch view := rawView.(type) {
	case resource.AppFileView:
		return t.viewAppFile(resources, view)
	case resource.EffectiveFileView:
		return t.viewEffectiveFile(resources, view)
	case resource.EffectiveResourceView:
		return t.viewEffectiveResource(resources, view)
	case resource.ValidateResourceView:
		return t.viewValidateResource(resources, view)
	default:
		return nil, fmt.Errorf("unsupported view: %T", rawView)
	}
}

func (t *translationJSON) UpdateResource(_ context.Context, all []resource.ResourceFile, fileToUpdate *resource.ResourceFile, data []byte) (*resource.ResourceFile, error) {
	path := fileToUpdate.Location.Path

	// Compute requestedLangTag
	matches := templateLanguageTagRegex.FindStringSubmatch(path)
	if len(matches) < 2 {
		return nil, resource.ErrResourceNotFound
	}
	requestedLangTag := matches[1]

	// View translation.json on all Fss but the app FS.
	// That is, the default translation.json.
	defaultTranslationObj := make(map[string]string)
	for _, r := range all {
		// Skip the app FS.
		if r.Location.Fs.GetFsLevel() == fileToUpdate.Location.Fs.GetFsLevel() {
			continue
		}

		// Skip file in different language.
		langTag := templateLanguageTagRegex.FindStringSubmatch(r.Location.Path)[1]
		if langTag != requestedLangTag {
			continue
		}

		var jsonObj map[string]interface{}
		err := json.Unmarshal(r.Data, &jsonObj)
		if err != nil {
			return nil, fmt.Errorf("translation file must be JSON: %w", err)
		}

		for key, val := range jsonObj {
			value, ok := val.(string)
			if !ok {
				return nil, fmt.Errorf("translation value must be string: %s %T", key, val)
			}
			defaultTranslationObj[key] = value
		}
	}

	appTranslationRaw := make(map[string]interface{})
	err := json.Unmarshal(data, &appTranslationRaw)
	if err != nil {
		return nil, fmt.Errorf("translation file must be JSON: %w", err)
	}
	appTranslationObj := make(map[string]string)
	for key, val := range appTranslationRaw {
		value, ok := val.(string)
		if !ok {
			return nil, fmt.Errorf("translation value must be string: %s %T", key, val)
		}
		appTranslationObj[key] = value
	}

	for key, val := range appTranslationObj {
		// If the value is the same as default, delete it.
		defaultValue := defaultTranslationObj[key]
		if defaultValue == val {
			delete(appTranslationObj, key)
		}

		// Note that we allow the app translation.json to contain unknown keys.
	}

	// If the translation is empty, delete the file instead.
	if len(appTranslationObj) <= 0 {
		data = nil
	} else {
		data, err = json.Marshal(appTranslationObj)
		if err != nil {
			return nil, err
		}
	}

	return &resource.ResourceFile{
		Location: fileToUpdate.Location,
		Data:     data,
	}, nil
}

func (t *translationJSON) viewValidateResource(resources []resource.ResourceFile, view resource.ValidateResourceView) (interface{}, error) {
	for _, resrc := range resources {
		langTag := templateLanguageTagRegex.FindStringSubmatch(resrc.Location.Path)[1]

		var jsonObj map[string]interface{}
		if err := json.Unmarshal(resrc.Data, &jsonObj); err != nil {
			return nil, fmt.Errorf("translation file must be JSON: %w", err)
		}

		for key, val := range jsonObj {
			value, ok := val.(string)
			if !ok {
				return nil, fmt.Errorf("translation `%v` must be string (%T)", key, val)
			}
			tag := language.Make(langTag)
			_, err := messageformat.FormatTemplateParseTree(tag, value)
			if err != nil {
				return nil, fmt.Errorf("translation `%v` is invalid: %w", key, err)
			}
		}
	}
	return nil, nil
}

func (t *translationJSON) viewEffectiveResource(resources []resource.ResourceFile, view resource.EffectiveResourceView) (interface{}, error) {
	type LanguageTag string
	type TranslationKey string
	type TranslationValue string

	preferredLanguageTags := view.PreferredLanguageTags()
	defaultLanguageTag := view.DefaultLanguageTag()

	appSpecificTranslationMap := make(map[TranslationKey]map[resource.FsLevel]map[LanguageTag]TranslationValue)
	translationMap := make(map[TranslationKey]map[LanguageTag]TranslationValue)

	add := func(langTag string, resrc resource.ResourceFile) error {
		var jsonObj map[string]interface{}
		if err := json.Unmarshal(resrc.Data, &jsonObj); err != nil {
			return fmt.Errorf("translation file must be JSON: %w", err)
		}

		fsLevel := resrc.Location.Fs.GetFsLevel()
		for key, val := range jsonObj {
			value, ok := val.(string)
			if !ok {
				return fmt.Errorf("translation `%v` must be string (%T)", key, val)
			}
			if t.IsAppSpecificKey(key) {
				// prepare app specific keys tanslation map
				keyTranslations, ok := appSpecificTranslationMap[TranslationKey(key)]
				if !ok {
					keyTranslations = make(map[resource.FsLevel]map[LanguageTag]TranslationValue)
					appSpecificTranslationMap[TranslationKey(key)] = keyTranslations
				}

				fsTranslations, ok := keyTranslations[fsLevel]
				if !ok {
					fsTranslations = make(map[LanguageTag]TranslationValue)
					keyTranslations[fsLevel] = fsTranslations
				}
				fsTranslations[LanguageTag(langTag)] = TranslationValue(value)
			} else {
				// prepare app agnostic keys tanslation map
				keyTranslations, ok := translationMap[TranslationKey(key)]
				if !ok {
					keyTranslations = make(map[LanguageTag]TranslationValue)
					translationMap[TranslationKey(key)] = keyTranslations
				}
				keyTranslations[LanguageTag(langTag)] = TranslationValue(value)
			}
		}
		return nil
	}
	extractLanguageTag := func(resrc resource.ResourceFile) string {
		langTag := templateLanguageTagRegex.FindStringSubmatch(resrc.Location.Path)[1]
		return langTag
	}

	err := intlresource.Prepare(resources, view, extractLanguageTag, add)
	if err != nil {
		return nil, err
	}

	translationData := make(map[string]Translation)
	// Prepare app agnostic data
	// We will first group all translations by the languages based on the fs level hierarchy
	// Higher fs level translations overwrite the lower one
	// After getting translations in all the languages
	// Resolve the translations bases on user's preferred language
	for key, translations := range translationMap {
		var items []intlresource.LanguageItem
		for languageTag, value := range translations {
			items = append(items, Translation{
				LanguageTag: string(languageTag),
				Value:       string(value),
			})
		}
		var matched intlresource.LanguageItem
		matched, err := intlresource.Match(preferredLanguageTags, defaultLanguageTag, items)
		if errors.Is(err, intlresource.ErrNoLanguageMatch) {
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

	// Preparing app specific data
	// If translations are provided in the higher fs level,
	// the translations will be resolved at that fs level.
	// Based on the user's preferred language, we will first look for the matched language,
	// the next will be the fallback language, the third will be any language other languages
	// We will only look for the translations from the lower fs level
	// if the keys are not provided in the higher fs level translations
	for key, translationsInFs := range appSpecificTranslationMap {
		for _, level := range fsLevelsOrderedInAscendingPriority {
			var items []intlresource.LanguageItem
			for translationFsLevel, translations := range translationsInFs {
				if level != translationFsLevel {
					continue
				}
				for languageTag, value := range translations {
					items = append(items, Translation{
						LanguageTag: string(languageTag),
						Value:       string(value),
					})
				}
			}

			var matched intlresource.LanguageItem
			matched, err := intlresource.Match(preferredLanguageTags, defaultLanguageTag, items)
			if errors.Is(err, intlresource.ErrNoLanguageMatch) {
				if len(items) > 0 {
					// Use first item in case of no match, to ensure resolution always succeed in the fs level
					matched = items[0]
				} else {
					// Ignore keys when no tranlations are provided in this fs level
					continue
				}
			} else if err != nil {
				return nil, err
			}
			translationData[string(key)] = matched.(Translation)
		}
	}

	// translationData
	return translationData, nil
}

func (t *translationJSON) viewAppFile(resources []resource.ResourceFile, view resource.AppFileView) (interface{}, error) {
	// AppFileView on translation.json returns the translation.json in the app FS if exists.
	path := view.AppFilePath()

	found := false
	var bytes []byte
	for _, resrc := range resources {
		if resrc.Location.Fs.GetFsLevel() == resource.FsLevelApp && path == resrc.Location.Path {
			found = true
			bytes = resrc.Data
		}
	}

	if !found {
		return nil, resource.ErrResourceNotFound
	}

	return bytes, nil
}

func (t *translationJSON) viewEffectiveFile(resources []resource.ResourceFile, view resource.EffectiveFileView) (interface{}, error) {
	// EffectiveFileView on translation.json is a simple merge
	// on the same file across different FSs.

	path := view.EffectiveFilePath()

	// Compute requestedLangTag
	matches := templateLanguageTagRegex.FindStringSubmatch(path)
	if len(matches) < 2 {
		return nil, resource.ErrResourceNotFound
	}
	requestedLangTag := matches[1]

	translationObj := make(map[string]string)
	for _, resrc := range resources {
		langTag := templateLanguageTagRegex.FindStringSubmatch(resrc.Location.Path)[1]

		if langTag == requestedLangTag {
			var jsonObj map[string]interface{}
			err := json.Unmarshal(resrc.Data, &jsonObj)
			if err != nil {
				return nil, fmt.Errorf("translation file must be JSON: %w", err)
			}

			for key, val := range jsonObj {
				value, ok := val.(string)
				if !ok {
					return nil, fmt.Errorf("translation value must be string: %s %T", key, val)
				}
				translationObj[key] = value
			}
		}
	}

	// As a special case, if the merged object is empty,
	// we report not found.
	if len(translationObj) <= 0 {
		return nil, resource.ErrResourceNotFound
	}

	// The effective file view is intended to be displayed to human for editing.
	// Therefore, we should disable HTML escape and add indentation.
	var buf bytes.Buffer
	encoder := json.NewEncoder(&buf)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "  ")
	err := encoder.Encode(translationObj)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
