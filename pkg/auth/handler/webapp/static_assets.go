package webapp

import (
	// nolint:gosec
	"crypto/md5"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path"
	"strings"

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
	Read(desc resource.Descriptor, view resource.View) (interface{}, error)
	Resolve(path string) (resource.Descriptor, bool)
}

type StaticAssetsHandler struct {
	Resources ResourceManager
}

func (h *StaticAssetsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fileServer := http.StripPrefix("/static/", http.FileServer(h))

	filePath := strings.TrimPrefix(r.URL.Path, "/static/")
	_, err := h.Open(filePath)
	if err == nil {
		// set cache control header if the file is found
		// 604800 seconds is a week
		w.Header().Set("Cache-Control", "public, max-age=604800")
	}

	fileServer.ServeHTTP(w, r)
}

func (h *StaticAssetsHandler) Open(name string) (http.File, error) {
	p := path.Join(web.StaticAssetResourcePrefix, name)

	filePath, hashInPath := web.ParsePathWithHash(p)
	if filePath == "" || hashInPath == "" {
		return nil, os.ErrNotExist
	}

	desc, ok := h.Resources.Resolve(filePath)
	if !ok {
		return nil, os.ErrNotExist
	}

	// We use EffectiveFile here because we want to return an exact match.
	// The static asset URLs in the templates are computed by the resolver using EffectiveResource, which has handled localization.
	result, err := h.Resources.Read(desc, resource.EffectiveFile{
		Path: filePath,
	})
	if errors.Is(err, resource.ErrResourceNotFound) {
		return nil, os.ErrNotExist
	} else if err != nil {
		return nil, err
	}

	bytes := result.([]byte)
	// check the hash
	// md5 is used to compute the hash in the filename for caching purpose only
	// nolint:gosec
	dataHash := md5.Sum(bytes)
	if fmt.Sprintf("%x", dataHash) != hashInPath {
		return nil, os.ErrNotExist
	}

	data := aferomem.CreateFile(p)
	file := aferomem.NewFileHandle(data)
	_, _ = file.Write(bytes)
	return aferomem.NewReadOnlyFileHandle(data), nil
}
