package handler

import (
	"errors"
	"net/http"
	"net/http/httputil"

	"github.com/gorilla/mux"

	"github.com/skygeario/skygear-server/pkg/core/cloudstorage"
	"github.com/skygeario/skygear-server/pkg/core/handler"
	coreHttp "github.com/skygeario/skygear-server/pkg/core/http"
	"github.com/skygeario/skygear-server/pkg/core/http/httpsigning"
	"github.com/skygeario/skygear-server/pkg/core/imageprocessing"
	"github.com/skygeario/skygear-server/pkg/core/inject"
	"github.com/skygeario/skygear-server/pkg/core/server"
	"github.com/skygeario/skygear-server/pkg/core/skyerr"
)

const (
	QueryNamePipeline = "pipeline"
)

var ErrBadAccess = errors.New("bad access")

func AttachGetHandler(
	server *server.Server,
	dependencyMap inject.DependencyMap,
) *server.Server {
	server.Handle("/get/{asset_name}", &GetHandlerFactory{
		dependencyMap,
	}).Methods("OPTIONS", "HEAD", "GET")
	return server
}

type GetHandlerFactory struct {
	DependencyMap inject.DependencyMap
}

func (f *GetHandlerFactory) NewHandler(request *http.Request) http.Handler {
	h := &GetHandler{}
	inject.DefaultRequestInject(h, f.DependencyMap, request)
	return h
}

/*
	@Operation GET /get/{asset_name} - Retrieve the asset
		Retrieve the asset.

		@Response 200
			The asset.
*/
type GetHandler struct {
	CloudStorageProvider cloudstorage.Provider `dependency:"CloudStorageProvider"`
}

func (h *GetHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	assetName := vars["asset_name"]

	isHead := r.Method == "HEAD"

	originallySigned := httpsigning.IsSigned(r)

	query := r.URL.Query()
	pipeline := query.Get(QueryNamePipeline)
	_, hasPipeline := query[QueryNamePipeline]
	query.Del(QueryNamePipeline)
	r.URL.RawQuery = query.Encode()

	if originallySigned {
		err := h.CloudStorageProvider.Verify(r)
		if err != nil {
			handler.WriteResponse(w, handler.APIResponse{
				Err: skyerr.MakeError(err),
			})
			return
		}
	}

	u, err := h.CloudStorageProvider.PresignGetRequest(assetName)
	if err != nil {
		handler.WriteResponse(w, handler.APIResponse{
			Err: skyerr.MakeError(err),
		})
		return
	}

	director := func(r *http.Request) {
		// Always set method to GET because S3 treats GET and HEAD differently.
		r.Method = "GET"
		// Remove irrelevant header.
		r.Header = coreHttp.RemoveSkygearHeader(r.Header)
		// Do not support range request if image processing query is present.
		if hasPipeline {
			r.Header.Del("Range")
			r.Header.Del("If-Range")
		}
		r.URL = u
		// Override the Host header
		r.Host = ""
		r.Header.Set("Host", u.Hostname())
	}

	modifyResponse := func(resp *http.Response) error {

		// We only know how to modify 2xx response.
		if resp.StatusCode < 200 || resp.StatusCode > 299 {
			return nil
		}

		resp.Header = h.CloudStorageProvider.ProprietaryToStandard(resp.Header)
		// Do not support range request if image processing query is present.
		if hasPipeline {
			resp.Header.Del("Accept-Ranges")
		}

		// Check access
		accessType := h.CloudStorageProvider.AccessType(resp.Header)
		if accessType == cloudstorage.AccessTypePrivate && !originallySigned {
			return ErrBadAccess
		}

		valid := imageprocessing.IsApplicableToHTTPResponse(resp)
		if isHead || !valid || !hasPipeline {
			return nil
		}
		ops, err := imageprocessing.Parse(pipeline)
		if err != nil {
			return nil
		}
		err = imageprocessing.ApplyToHTTPResponse(resp, ops)
		if err != nil {
			return err
		}

		return nil
	}

	errorHandler := func(w http.ResponseWriter, req *http.Request, err error) {
		if err == ErrBadAccess {
			w.WriteHeader(http.StatusUnauthorized)
		} else {
			w.WriteHeader(http.StatusBadGateway)
		}
	}

	reverseProxy := &httputil.ReverseProxy{
		Director:       director,
		ModifyResponse: modifyResponse,
		ErrorHandler:   errorHandler,
	}

	reverseProxy.ServeHTTP(w, r)
}
