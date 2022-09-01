package webapp

import (
	// nolint:gosec
	"net/http"
	"os"
	"path"

	"github.com/authgear/authgear-server/pkg/lib/web"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/httputil"
)

func ConfigureGeneratedStaticAssetsRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("HEAD", "GET", "OPTIONS").
		WithPathPattern("/generated/*all")
}

type GlobalEmbeddedResourceManager interface {
	Resolve(resourcePath string) (string, bool)
	Open(assetPath string) (http.File, error)
}

type GeneratedStaticAssetsHandler struct {
	EmbeddedResources GlobalEmbeddedResourceManager
}

func (h *GeneratedStaticAssetsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fileServer := http.StripPrefix("/generated/", &httputil.FileServer{
		FileSystem:          h,
		FallbackToIndexHTML: false,
	})
	fileServer.ServeHTTP(w, r)
}

func (h *GeneratedStaticAssetsHandler) Open(name string) (http.File, error) {
	p := path.Join(web.DefaultResourcePrefix, name)
	if asset, ok := h.EmbeddedResources.Resolve(p); ok {
		return h.EmbeddedResources.Open(asset)
	}
	return nil, os.ErrNotExist
}
