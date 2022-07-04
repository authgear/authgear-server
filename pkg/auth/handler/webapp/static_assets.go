package webapp

import (
	"bytes"
	// nolint:gosec
	"crypto/md5"
	"errors"
	"fmt"
	"mime"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/authgear/authgear-server/pkg/lib/web"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/readcloserthunk"
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
	filePath := strings.TrimPrefix(r.URL.Path, "/static/")
	thunk, err := h.GetThunk(filePath)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	// set cache control header if the file is found
	// 604800 seconds is a week
	w.Header().Set("Cache-Control", "public, max-age=604800")

	w.Header().Set("Content-Type", mime.TypeByExtension(path.Ext(filePath)))
	_, err = readcloserthunk.Copy(w, thunk)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *StaticAssetsHandler) GetThunk(name string) (readcloserthunk.ReadCloserThunk, error) {
	p := path.Join(web.StaticAssetResourcePrefix, name)

	filePath, hashInPath := web.ParsePathWithHash(p)
	if filePath == "" || hashInPath == "" {
		return nil, os.ErrNotExist
	}

	resolvePath := filePath

	desc, ok := h.Resources.Resolve(resolvePath)
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

	// Given that there are different output types from different descriptors with EffectiveFile view
	switch v := result.(type) {
	// every generated asset always output ReadCloserThunk
	case readcloserthunk.ReadCloserThunk:
		return v, nil
	// other asset output []byte
	case []byte:
		// check the hash
		// md5 is used to compute the hash in the filename for caching purpose only
		// nolint:gosec
		dataHash := md5.Sum(v)
		if fmt.Sprintf("%x", dataHash) != hashInPath {
			return nil, os.ErrNotExist
		}
		return readcloserthunk.Reader(bytes.NewReader(v)), nil
	}

	return nil, os.ErrNotExist
}
