package webapp

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/lib/web"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/httputil"
)

func ConfigureGeneratedStaticAssetsRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("HEAD", "GET", "OPTIONS").
		WithPathPattern("/" + web.GeneratedAssetsURLDirname + "/*all")
}

type GlobalEmbeddedResourceManager interface {
	Open(name string) (http.File, error)
}

type GeneratedStaticAssetsHandler struct {
	EmbeddedResources GlobalEmbeddedResourceManager
}

func (h *GeneratedStaticAssetsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fileServer := &httputil.FileServer{
		FileSystem:          h.EmbeddedResources,
		AssetsDir:           web.GeneratedAssetsURLDirname,
		FallbackToIndexHTML: false,
	}
	fileServer.ServeHTTP(w, r)
}
