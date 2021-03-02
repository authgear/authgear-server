package intlresource

import (
	"github.com/authgear/authgear-server/pkg/util/intl"
	"github.com/authgear/authgear-server/pkg/util/resource"
)

// Prepare ensures that the supported languages always include intl.DefaultLanguage.
func Prepare(
	resources []resource.ResourceFile,
	view resource.EffectiveResourceView,
	extractLanguageTag func(resrc resource.ResourceFile) string,
	add func(langTag string, resrc resource.ResourceFile) error,
) error {
	supportedLanguageTags := view.SupportedLanguageTags()

	supportedSet := make(map[string]struct{})
	for _, tag := range supportedLanguageTags {
		supportedSet[tag] = struct{}{}
	}
	// Include intl.DefaultLanguage unconditionally.
	supportedSet[intl.DefaultLanguage] = struct{}{}

	for _, resrc := range resources {
		langTag := extractLanguageTag(resrc)

		// Ignore resources in unsupported languages.
		_, supported := supportedSet[langTag]
		if !supported {
			continue
		}

		err := add(langTag, resrc)
		if err != nil {
			return err
		}
	}
	return nil
}
