package template

import (
	htmltemplate "html/template"
	texttemplate "text/template"

	"github.com/authgear/authgear-server/pkg/util/resource"
)

type DefaultTemplateLanguage string

type ResourceManager interface {
	Read(desc resource.Descriptor, view resource.View) (interface{}, error)
}

type Resolver struct {
	Resources          ResourceManager
	DefaultLanguageTag DefaultTemplateLanguage
}

func (r *Resolver) ResolveHTML(desc *HTML, preferredLanguages []string) (*htmltemplate.Template, error) {
	resrc, err := r.Resources.Read(desc, resource.EffectiveResource{
		PreferredTags: preferredLanguages,
		DefaultTag:    string(r.DefaultLanguageTag),
	})
	if err != nil {
		return nil, err
	}

	return resrc.(*htmltemplate.Template), nil
}

func (r *Resolver) ResolvePlainText(desc *PlainText, preferredLanguages []string) (*texttemplate.Template, error) {
	resrc, err := r.Resources.Read(desc, resource.EffectiveResource{
		PreferredTags: preferredLanguages,
		DefaultTag:    string(r.DefaultLanguageTag),
	})
	if err != nil {
		return nil, err
	}

	return resrc.(*texttemplate.Template), nil
}

func (r *Resolver) ResolveTranslations(preferredLanguages []string) (map[string]Translation, error) {
	resrc, err := r.Resources.Read(TranslationJSON, resource.EffectiveResource{
		PreferredTags: preferredLanguages,
		DefaultTag:    string(r.DefaultLanguageTag),
	})
	if err != nil {
		return nil, err
	}

	return resrc.(map[string]Translation), nil
}
