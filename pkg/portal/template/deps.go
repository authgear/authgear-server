package template

import (
	"github.com/google/wire"
	"github.com/spf13/afero"

	"github.com/authgear/authgear-server/pkg/util/fs"
	"github.com/authgear/authgear-server/pkg/util/template"
)

func NewEngine(defaultDir string) *template.Engine {
	// An empty FS to fill the requirement.
	fs := &fs.AferoFs{Fs: afero.NewMemMapFs()}
	resolver := template.NewResolver(template.NewResolverOptions{
		AppFs:                     fs,
		Registry:                  template.DefaultRegistry.Clone(),
		DefaultTemplatesDirectory: defaultDir,
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
