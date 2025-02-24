package webapp

import (
	"context"
	// nolint:gosec
	"crypto/md5"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"

	aferomem "github.com/spf13/afero/mem"

	"github.com/authgear/authgear-server/pkg/lib/web"
	"github.com/authgear/authgear-server/pkg/util/filepathutil"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/httputil"
	"github.com/authgear/authgear-server/pkg/util/resource"
)

func ConfigureAppStaticAssetsRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("HEAD", "GET").
		WithPathPattern("/" + web.AppAssetsURLDirname + "/*all")
}

type ResourceManager interface {
	Read(ctx context.Context, desc resource.Descriptor, view resource.View) (interface{}, error)
	Resolve(path string) (resource.Descriptor, bool)
}

type AppStaticAssetsHandler struct {
	Resources ResourceManager
}

func (h *AppStaticAssetsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fileServer := &httputil.FileServer{
		FileSystem:          h.makeFs(r.Context()),
		AssetsDir:           web.AppAssetsURLDirname,
		FallbackToIndexHTML: false,
	}
	fileServer.ServeHTTP(w, r)
}

func (h *AppStaticAssetsHandler) makeFs(ctx context.Context) http.FileSystem {
	return &handlerFs{
		ctx:     ctx,
		handler: h,
	}
}

type handlerFs struct {
	ctx     context.Context
	handler *AppStaticAssetsHandler
}

func (f *handlerFs) Open(name string) (http.File, error) {
	ctx := f.ctx

	// ResourceManager.Resolve does not expect a leading slash.
	p := strings.TrimPrefix(name, "/")

	filePath, hashInPath, ok := filepathutil.ParseHashedPath(p)
	if !ok {
		return nil, os.ErrNotExist
	}

	// Fallback ResourceManager
	desc, ok := f.handler.Resources.Resolve(filePath)
	if !ok {
		return nil, os.ErrNotExist
	}

	// We use EffectiveFile here because we want to return an exact match.
	// The static asset URLs in the templates are computed by the resolver using EffectiveResource, which has handled localization.
	result, err := f.handler.Resources.Read(ctx, desc, resource.EffectiveFile{
		Path: filePath,
	})
	if errors.Is(err, resource.ErrResourceNotFound) {
		return nil, os.ErrNotExist
	} else if err != nil {
		return nil, err
	}

	bytes := result.([]byte)
	if !web.LookLikeAHash(hashInPath) {
		// check the hash
		// md5 is used to compute the hash in the filename for caching purpose only
		// nolint:gosec
		dataHash := md5.Sum(bytes)
		if fmt.Sprintf("%x", dataHash) != hashInPath {
			return nil, os.ErrNotExist
		}
	}

	data := aferomem.CreateFile(p)
	file := aferomem.NewFileHandle(data)
	_, _ = file.Write(bytes)
	return aferomem.NewReadOnlyFileHandle(data), nil
}

var _ http.FileSystem = &handlerFs{}
