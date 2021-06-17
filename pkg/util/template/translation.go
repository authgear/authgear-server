package template

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"

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

type translationJSON struct{}

var _ resource.Descriptor = &translationJSON{}

var TranslationJSON = resource.RegisterResource(&translationJSON{})

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

func (t *translationJSON) UpdateResource(resrc *resource.ResourceFile, data []byte, view resource.View) (*resource.ResourceFile, error) {
	return &resource.ResourceFile{
		Location: resrc.Location,
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

	translationMap := make(map[TranslationKey]map[LanguageTag]TranslationValue)
	add := func(langTag string, resrc resource.ResourceFile) error {
		var jsonObj map[string]interface{}
		if err := json.Unmarshal(resrc.Data, &jsonObj); err != nil {
			return fmt.Errorf("translation file must be JSON: %w", err)
		}

		for key, val := range jsonObj {
			value, ok := val.(string)
			if !ok {
				return fmt.Errorf("translation `%v` must be string (%T)", key, val)
			}
			keyTranslations, ok := translationMap[TranslationKey(key)]
			if !ok {
				keyTranslations = make(map[LanguageTag]TranslationValue)
				translationMap[TranslationKey(key)] = keyTranslations
			}
			keyTranslations[LanguageTag(langTag)] = TranslationValue(value)
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
