package deps

import (
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/fs"
	"github.com/authgear/authgear-server/pkg/util/template"
)

func NewEngineWithConfig(
	appFs fs.Fs,
	defaultDir string,
	c *config.Config,
) *template.Engine {
	var refs []template.Reference
	for _, item := range c.AppConfig.Template.Items {
		refs = append(refs, template.Reference{
			Type:        string(item.Type),
			LanguageTag: item.LanguageTag,
			URI:         item.URI,
		})
	}

	resolver := template.NewResolver(template.NewResolverOptions{
		AppFs:                     appFs,
		Registry:                  template.DefaultRegistry.Clone(),
		DefaultTemplatesDirectory: defaultDir,
		References:                refs,
		FallbackLanguageTag:       c.AppConfig.Localization.FallbackLanguage,
	})
	engine := &template.Engine{
		Resolver: resolver,
	}

	return engine
}
