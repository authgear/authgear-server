package deps

import (
	"github.com/spf13/afero"

	"github.com/authgear/authgear-server/pkg/util/resource"
)

func NewResourceManager(builtinResourceDir string, customResourceDir string) *resource.Manager {
	var fs []resource.Fs
	fs = append(fs,
		resource.AferoFs{Fs: afero.NewBasePathFs(afero.OsFs{}, builtinResourceDir)},
	)
	if customResourceDir != "" {
		fs = append(fs,
			resource.AferoFs{Fs: afero.NewBasePathFs(afero.OsFs{}, customResourceDir)},
		)
	}
	return &resource.Manager{
		Registry: resource.DefaultRegistry.Clone(),
		Fs:       fs,
	}
}
