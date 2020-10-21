package webapp

import (
	"errors"
	"net/http"
	"os"
	"path"

	aferomem "github.com/spf13/afero/mem"

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
	Read(desc resource.Descriptor, args map[string]interface{}) (*resource.MergedFile, error)
	Resolve(path string) (resource.Descriptor, bool)
}

type StaticAssetsHandler struct {
	Resources ResourceManager
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

	merged, err := h.Resources.Read(desc, nil)
	if errors.Is(err, resource.ErrResourceNotFound) {
		return nil, os.ErrNotExist
	} else if err != nil {
		return nil, err
	}

	asset, err := desc.Parse(merged)
	if err != nil {
		return nil, err
	}

	sAsset := asset.(*web.StaticAsset)
	if sAsset.Path != p {
		return nil, os.ErrNotExist
	}

	data := aferomem.CreateFile(sAsset.Path)
	file := aferomem.NewFileHandle(data)
	_, _ = file.Write(sAsset.Data)
	return aferomem.NewReadOnlyFileHandle(data), nil
}
