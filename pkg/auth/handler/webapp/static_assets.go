package webapp

import (
	"errors"
	"net/http"
	"os"
	"path"

	aferomem "github.com/spf13/afero/mem"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/web"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/resource"
)

func ConfigureStaticAssetsRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("HEAD", "GET").
		WithPathPattern("/static/*all")
}

type ResourceManager interface {
	Read(desc resource.Descriptor, view resource.View) (interface{}, error)
	Resolve(path string) (resource.Descriptor, bool)
}

type StaticAssetsHandler struct {
	Resources    ResourceManager
	Localization *config.LocalizationConfig
}

func (h *StaticAssetsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fileServer := http.StripPrefix("/static/", http.FileServer(h))
	fileServer.ServeHTTP(w, r)
}

func (h *StaticAssetsHandler) Open(name string) (http.File, error) {
	p := path.Join(web.StaticAssetResourcePrefix, name)

	desc, ok := h.Resources.Resolve(p)
	if !ok {
		return nil, os.ErrNotExist
	}

	// We use EffectiveFile here because we want to return an exact match.
	// The static asset URLs in the templates are computed by the resolver using EffectiveResource, which has handled localization.
	result, err := h.Resources.Read(desc, resource.EffectiveFile{
		Path:       p,
		DefaultTag: h.Localization.FallbackLanguage,
	})
	if errors.Is(err, resource.ErrResourceNotFound) {
		return nil, os.ErrNotExist
	} else if err != nil {
		return nil, err
	}

	bytes := result.([]byte)
	data := aferomem.CreateFile(p)
	file := aferomem.NewFileHandle(data)
	_, _ = file.Write(bytes)
	return aferomem.NewReadOnlyFileHandle(data), nil
}
