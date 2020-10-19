package deps

import (
	"github.com/spf13/afero"

	"github.com/authgear/authgear-server/pkg/util/resource"
)

func NewResourceManager(defaultResourceDir string) *resource.Manager {
	return &resource.Manager{
		Registry: resource.DefaultRegistry.Clone(),
		Fs: []resource.Fs{
			resource.AferoFs{Fs: afero.NewBasePathFs(afero.OsFs{}, defaultResourceDir)},
		},
	}
}
