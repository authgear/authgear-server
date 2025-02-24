package template

import (
	"context"

	"github.com/authgear/authgear-server/pkg/util/resource"
)

type DefaultLanguageTag string
type SupportedLanguageTags []string

type ResourceManager interface {
	Read(ctx context.Context, desc resource.Descriptor, view resource.View) (interface{}, error)
}

type Resolver struct {
	Resources             ResourceManager
	DefaultLanguageTag    DefaultLanguageTag
	SupportedLanguageTags SupportedLanguageTags
}

func (r *Resolver) ResolveHTML(ctx context.Context, desc *HTML, preferredLanguages []string) (*HTMLTemplateEffectiveResource, error) {
	resrc, err := r.Resources.Read(ctx, desc, resource.EffectiveResource{
		SupportedTags: []string(r.SupportedLanguageTags),
		DefaultTag:    string(r.DefaultLanguageTag),
		PreferredTags: preferredLanguages,
	})
	if err != nil {
		return nil, err
	}

	return resrc.(*HTMLTemplateEffectiveResource), nil
}

func (r *Resolver) ResolveMessageHTML(ctx context.Context, desc *MessageHTML, preferredLanguages []string) (*HTMLTemplateEffectiveResource, error) {
	resrc, err := r.Resources.Read(ctx, desc, resource.EffectiveResource{
		SupportedTags: []string(r.SupportedLanguageTags),
		DefaultTag:    string(r.DefaultLanguageTag),
		PreferredTags: preferredLanguages,
	})
	if err != nil {
		return nil, err
	}

	return resrc.(*HTMLTemplateEffectiveResource), nil
}

func (r *Resolver) ResolvePlainText(ctx context.Context, desc *PlainText, preferredLanguages []string) (*TextTemplateEffectiveResource, error) {
	resrc, err := r.Resources.Read(ctx, desc, resource.EffectiveResource{
		SupportedTags: []string(r.SupportedLanguageTags),
		DefaultTag:    string(r.DefaultLanguageTag),
		PreferredTags: preferredLanguages,
	})
	if err != nil {
		return nil, err
	}

	return resrc.(*TextTemplateEffectiveResource), nil
}

func (r *Resolver) ResolveMessagePlainText(ctx context.Context, desc *MessagePlainText, preferredLanguages []string) (*TextTemplateEffectiveResource, error) {
	resrc, err := r.Resources.Read(ctx, desc, resource.EffectiveResource{
		SupportedTags: []string(r.SupportedLanguageTags),
		DefaultTag:    string(r.DefaultLanguageTag),
		PreferredTags: preferredLanguages,
	})
	if err != nil {
		return nil, err
	}

	return resrc.(*TextTemplateEffectiveResource), nil
}

func (r *Resolver) ResolveTranslations(ctx context.Context, preferredLanguages []string) (map[string]Translation, error) {
	resrc, err := r.Resources.Read(ctx, TranslationJSON, resource.EffectiveResource{
		SupportedTags: []string(r.SupportedLanguageTags),
		DefaultTag:    string(r.DefaultLanguageTag),
		PreferredTags: preferredLanguages,
	})
	if err != nil {
		return nil, err
	}

	return resrc.(map[string]Translation), nil
}
