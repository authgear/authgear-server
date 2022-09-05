package webapp

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/httputil"
)

func ConfigureGeneratedStaticAssetsRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("HEAD", "GET", "OPTIONS").
		WithPathPattern("/generated/*all")
}

type GlobalEmbeddedResourceManager interface {
	Open(name string) (http.File, error)
}

type GeneratedStaticAssetsHandler struct {
	EmbeddedResources GlobalEmbeddedResourceManager
}

func (h *GeneratedStaticAssetsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fileServer := http.StripPrefix("/generated/", &httputil.FileServer{
		FileSystem:          h.EmbeddedResources,
		FallbackToIndexHTML: false,
	})
	fileServer.ServeHTTP(w, r)
}
