package template

import (
	htmltemplate "html/template"
	texttemplate "text/template"

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

func (r *Resolver) ResolveHTML(desc *HTML, preferredLanguages []string) (*htmltemplate.Template, error) {
	resrc, err := r.Resources.Read(desc, resource.EffectiveResource{
		SupportedTags: []string(r.SupportedLanguageTags),
		DefaultTag:    string(r.DefaultLanguageTag),
		PreferredTags: preferredLanguages,
	})
	if err != nil {
		return nil, err
	}

	return resrc.(*htmltemplate.Template), nil
}

func (r *Resolver) ResolveMessageHTML(desc *MessageHTML, preferredLanguages []string) (*htmltemplate.Template, error) {
	resrc, err := r.Resources.Read(desc, resource.EffectiveResource{
		SupportedTags: []string(r.SupportedLanguageTags),
		DefaultTag:    string(r.DefaultLanguageTag),
		PreferredTags: preferredLanguages,
	})
	if err != nil {
		return nil, err
	}

	return resrc.(*htmltemplate.Template), nil
}

func (r *Resolver) ResolvePlainText(desc *PlainText, preferredLanguages []string) (*texttemplate.Template, error) {
	resrc, err := r.Resources.Read(desc, resource.EffectiveResource{
		SupportedTags: []string(r.SupportedLanguageTags),
		DefaultTag:    string(r.DefaultLanguageTag),
		PreferredTags: preferredLanguages,
	})
	if err != nil {
		return nil, err
	}

	return resrc.(*texttemplate.Template), nil
}

func (r *Resolver) ResolveMessagePlainText(desc *MessagePlainText, preferredLanguages []string) (*texttemplate.Template, error) {
	resrc, err := r.Resources.Read(desc, resource.EffectiveResource{
		SupportedTags: []string(r.SupportedLanguageTags),
		DefaultTag:    string(r.DefaultLanguageTag),
		PreferredTags: preferredLanguages,
	})
	if err != nil {
		return nil, err
	}

	return resrc.(*texttemplate.Template), nil
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
