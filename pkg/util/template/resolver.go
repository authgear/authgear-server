package template

import (
	"github.com/authgear/authgear-server/pkg/util/resource"
)

type DefaultLanguageTag string
type SupportedLanguageTags []string

type ResourceManager interface {
	Read(desc resource.Descriptor, view resource.View) (interface{}, error)
}

type Resolver struct {
	Resources             ResourceManager
	DefaultLanguageTag    DefaultLanguageTag
	SupportedLanguageTags SupportedLanguageTags
}

func (r *Resolver) ResolveHTML(desc *HTML, preferredLanguages []string) (*HTMLTemplateEffectiveResource, error) {
	resrc, err := r.Resources.Read(desc, resource.EffectiveResource{
		SupportedTags: []string(r.SupportedLanguageTags),
		DefaultTag:    string(r.DefaultLanguageTag),
		PreferredTags: preferredLanguages,
	})
	if err != nil {
		return nil, err
	}

	return resrc.(*HTMLTemplateEffectiveResource), nil
}

func (r *Resolver) ResolveMessageHTML(desc *MessageHTML, preferredLanguages []string) (*HTMLTemplateEffectiveResource, error) {
	resrc, err := r.Resources.Read(desc, resource.EffectiveResource{
		SupportedTags: []string(r.SupportedLanguageTags),
		DefaultTag:    string(r.DefaultLanguageTag),
		PreferredTags: preferredLanguages,
	})
	if err != nil {
		return nil, err
	}

	return resrc.(*HTMLTemplateEffectiveResource), nil
}

func (r *Resolver) ResolvePlainText(desc *PlainText, preferredLanguages []string) (*TextTemplateEffectiveResource, error) {
	resrc, err := r.Resources.Read(desc, resource.EffectiveResource{
		SupportedTags: []string(r.SupportedLanguageTags),
		DefaultTag:    string(r.DefaultLanguageTag),
		PreferredTags: preferredLanguages,
	})
	if err != nil {
		return nil, err
	}

	return resrc.(*TextTemplateEffectiveResource), nil
}

func (r *Resolver) ResolveMessagePlainText(desc *MessagePlainText, preferredLanguages []string) (*TextTemplateEffectiveResource, error) {
	resrc, err := r.Resources.Read(desc, resource.EffectiveResource{
		SupportedTags: []string(r.SupportedLanguageTags),
		DefaultTag:    string(r.DefaultLanguageTag),
		PreferredTags: preferredLanguages,
	})
	if err != nil {
		return nil, err
	}

	return resrc.(*TextTemplateEffectiveResource), nil
}

func (r *Resolver) ResolveTranslations(preferredLanguages []string) (map[string]Translation, error) {
	resrc, err := r.Resources.Read(TranslationJSON, resource.EffectiveResource{
		SupportedTags: []string(r.SupportedLanguageTags),
		DefaultTag:    string(r.DefaultLanguageTag),
		PreferredTags: preferredLanguages,
	})
	if err != nil {
		return nil, err
	}

	return resrc.(map[string]Translation), nil
}
