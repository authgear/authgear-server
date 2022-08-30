package webapp

import (
	// nolint:gosec
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/authgear/authgear-server/pkg/lib/web"
	"github.com/authgear/authgear-server/pkg/util/httproute"
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
	fileServer := http.StripPrefix("/generated/", http.FileServer(h))

	filePath := strings.TrimPrefix(r.URL.Path, "/generated/")
	ok := h.Resolve(filePath)
	if ok {
		// set cache control header if the file is found
		// 604800 seconds is a week
		w.Header().Set("Cache-Control", "public, max-age=604800")
	}

	fileServer.ServeHTTP(w, r)
}

func (h *GeneratedStaticAssetsHandler) Resolve(name string) bool {
	p := path.Join(web.DefaultResourcePrefix, name)
	if _, ok := h.EmbeddedResources.Resolve(p); ok {
		return true
	}
	return false
}

func (h *GeneratedStaticAssetsHandler) Open(name string) (http.File, error) {
	p := path.Join(web.DefaultResourcePrefix, name)
	if asset, ok := h.EmbeddedResources.Resolve(p); ok {
		return h.EmbeddedResources.Open(asset)
	}
	return nil, os.ErrNotExist
}
