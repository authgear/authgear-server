package intlresource

import (
	"github.com/authgear/authgear-server/pkg/util/intl"
	"github.com/authgear/authgear-server/pkg/util/resource"
)

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

	var fallbackResrc resource.ResourceFile
	added := false
	for _, resrc := range resources {
		langTag := extractLanguageTag(resrc)
		if langTag == intl.DefaultLanguage {
			fallbackResrc = resrc
		}
		// Ignore resources in unsupported languages.
		_, supported := supportedSet[langTag]
		if !supported {
			continue
		}
		err := add(langTag, resrc)
		if err != nil {
			return err
		}
		added = true
	}
	// Add fallback resource.
	if !added {
		err := add(intl.DefaultLanguage, fallbackResrc)
		if err != nil {
			return err
		}
	}
	return nil
}
