package template

import (
	htmltemplate "html/template"
	texttemplate "text/template"

	"github.com/authgear/authgear-server/pkg/util/resource"
	"github.com/authgear/authgear-server/pkg/util/template"
)

type DefaultTemplateLanguage string

type ResourceManager interface {
	Read(desc resource.Descriptor, args map[string]interface{}) (*resource.LayerFile, error)
}

type Resolver struct {
	Resources          ResourceManager
	DefaultLanguageTag DefaultTemplateLanguage
}

func (r *Resolver) ResolveHTML(desc *HTML, preferredLanguages []string) (*htmltemplate.Template, error) {
	file, err := r.Resources.Read(desc, map[string]interface{}{
		ResourceArgPreferredLanguageTag: preferredLanguages,
		ResourceArgDefaultLanguageTag:   string(r.DefaultLanguageTag),
	})
	if err != nil {
		return nil, err
	}

	tpl, err := desc.Parse(file.Data)
	if err != nil {
		return nil, err
	}

	return tpl.(*htmltemplate.Template), nil
}

func (r *Resolver) ResolvePlainText(desc *PlainText, preferredLanguages []string) (*texttemplate.Template, error) {
	file, err := r.Resources.Read(desc, map[string]interface{}{
		ResourceArgPreferredLanguageTag: preferredLanguages,
		ResourceArgDefaultLanguageTag:   string(r.DefaultLanguageTag),
	})
	if err != nil {
		return nil, err
	}

	tpl, err := desc.Parse(file.Data)
	if err != nil {
		return nil, err
	}

	return tpl.(*texttemplate.Template), nil
}

func (r *Resolver) ResolveTranslations(preferredLanguages []string) (map[string]template.Translation, error) {
	file, err := r.Resources.Read(TranslationJSON, map[string]interface{}{
		ResourceArgPreferredLanguageTag: preferredLanguages,
		ResourceArgDefaultLanguageTag:   string(r.DefaultLanguageTag),
	})
	if err != nil {
		return nil, err
	}

	ts, err := TranslationJSON.Parse(file.Data)
	if err != nil {
		return nil, err
	}

	return ts.(map[string]template.Translation), nil
}
