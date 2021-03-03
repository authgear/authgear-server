package intlresource

import (
	"github.com/authgear/authgear-server/pkg/util/intl"
	"github.com/authgear/authgear-server/pkg/util/resource"
)

// PrepareNonFallback prepares resources in non-fallback languages.
func PrepareNonFallback(
	resources []resource.ResourceFile,
	view resource.EffectiveResourceView,
	extractLanguageTag func(resrc resource.ResourceFile) string,
	add func(langTag string, resrc resource.ResourceFile) error,
) error {
	supportedLanguageTags := view.SupportedLanguageTags()
	defaultLanguageTag := view.DefaultLanguageTag()

	nonFallbackSet := make(map[string]struct{})
	for _, tag := range supportedLanguageTags {
		nonFallbackSet[tag] = struct{}{}
	}
	// Exclude the fallback language.
	delete(nonFallbackSet, defaultLanguageTag)

	for _, resrc := range resources {
		langTag := extractLanguageTag(resrc)

		_, isNonFallback := nonFallbackSet[langTag]
		if !isNonFallback {
			continue
		}

		err := add(langTag, resrc)
		if err != nil {
			return err
		}
	}
	return nil
}

// PrepareFallback prepares resources in fallback language.
func PrepareFallback(
	resources []resource.ResourceFile,
	view resource.EffectiveResourceView,
	extractLanguageTag func(resrc resource.ResourceFile) string,
	add func(langTag string, resrc resource.ResourceFile) error,
) error {
	defaultLanguageTag := view.DefaultLanguageTag()

	// Add the builtin resource of intl.DefaultLanguage first.
	for _, resrc := range resources {
		langTag := extractLanguageTag(resrc)
		if resrc.Location.Fs.GetFsLevel() == resource.FsLevelBuiltin && langTag == intl.DefaultLanguage {
			err := add(defaultLanguageTag, resrc)
			if err != nil {
				return err
			}
		}
	}

	// Add the resources of fallback language.
	for _, resrc := range resources {
		langTag := extractLanguageTag(resrc)
		if langTag == defaultLanguageTag {
			err := add(defaultLanguageTag, resrc)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// Prepare is PrepareFallback, followed by PrepareNonFallback.
// Prepare ensures that the fallback resource is full and complete,
// by using the builtin one as the base.
func Prepare(
	resources []resource.ResourceFile,
	view resource.EffectiveResourceView,
	extractLanguageTag func(resrc resource.ResourceFile) string,
	add func(langTag string, resrc resource.ResourceFile) error,
) error {
	err := PrepareFallback(resources, view, extractLanguageTag, add)
	if err != nil {
		return err
	}
	err = PrepareNonFallback(resources, view, extractLanguageTag, add)
	if err != nil {
		return err
	}
	return nil
}
