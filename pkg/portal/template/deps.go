package template

import (
	"github.com/google/wire"
	"github.com/spf13/afero"

	portalconfig "github.com/authgear/authgear-server/pkg/portal/config"
	"github.com/authgear/authgear-server/pkg/util/fs"
	"github.com/authgear/authgear-server/pkg/util/template"
)

func NewEngine(defaultDir portalconfig.DefaultTemplateDirectory) *template.Engine {
	// An empty FS to fill the requirement.
	fs := &fs.AferoFs{Fs: afero.NewMemMapFs()}
	resolver := template.NewResolver(template.NewResolverOptions{
		AppFs:                     fs,
		Registry:                  template.DefaultRegistry.Clone(),
		DefaultTemplatesDirectory: string(defaultDir),
		// References and FallbackLanguageTag can be left as zero values.
	})
	engine := &template.Engine{
		Resolver: resolver,
	}
	return engine
}

var DependencySet = wire.NewSet(
	NewEngine,
)
