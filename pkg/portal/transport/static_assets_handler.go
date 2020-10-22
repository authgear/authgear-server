package transport

import (
	"net/http"
	"os"
	"path"

	"github.com/authgear/authgear-server/pkg/util/httputil"
	"github.com/authgear/authgear-server/pkg/util/resource"
)

type ResourceManager interface {
	Filesystems() []resource.Fs
}

type StaticAssetsHandler struct {
	Resources ResourceManager
}

func (h *StaticAssetsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	server := http.FileServer(&httputil.TryFileSystem{
		Fallback: "/index.html",
		FS:       h,
	})
	server.ServeHTTP(w, r)
}

func (h *StaticAssetsHandler) Open(name string) (http.File, error) {
	assetPath := path.Join("static", name)

	var effectiveFs resource.Fs
	for _, fs := range h.Resources.Filesystems() {
		_, err := fs.Stat(assetPath)
		if os.IsNotExist(err) {
			continue
		} else if err != nil {
			return nil, err
		}

		effectiveFs = fs
	}

	if effectiveFs == nil {
		return nil, os.ErrNotExist
	}

	return effectiveFs.Open(assetPath)
}
