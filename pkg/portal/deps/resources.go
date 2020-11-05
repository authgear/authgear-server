package deps

import (
	"github.com/spf13/afero"

	portalresource "github.com/authgear/authgear-server/pkg/portal/resource"
	"github.com/authgear/authgear-server/pkg/util/resource"
)

type AppBaseResources *resource.Manager

func ProvideAppBaseResources(root *RootProvider) AppBaseResources {
	return root.AppBaseResources
}

func NewAppResourceManager(builtinResourceDir string, customResourceDir string) *resource.Manager {
	var fs []resource.Fs
	fs = append(fs,
		resource.AferoFs{Fs: afero.NewBasePathFs(afero.OsFs{}, builtinResourceDir)},
	)
	if customResourceDir != "" {
		fs = append(fs,
			resource.AferoFs{Fs: afero.NewBasePathFs(afero.OsFs{}, customResourceDir)},
		)
	}
	return AppBaseResources(&resource.Manager{
		Registry: resource.DefaultRegistry.Clone(),
		Fs:       fs,
	})
}

func NewPortalResourceManager(builtinResourceDir string, customResourceDir string) *resource.Manager {
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
		Registry: portalresource.PortalRegistry.Clone(),
		Fs:       fs,
	}
}
