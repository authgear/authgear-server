package handler

import (
	"errors"
	"net/http"
	"net/http/httputil"

	"github.com/gorilla/mux"

	"github.com/skygeario/skygear-server/pkg/core/cloudstorage"
	coreHttp "github.com/skygeario/skygear-server/pkg/core/http"
	"github.com/skygeario/skygear-server/pkg/core/imageprocessing"
	"github.com/skygeario/skygear-server/pkg/core/inject"
	"github.com/skygeario/skygear-server/pkg/core/server"
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

type GetHandler struct {
	CloudStorageProvider cloudstorage.Provider `dependency:"CloudStorageProvider"`
}

func (h *GetHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	originallySigned := false
	pipeline := ""
	hasPipeline := false
	vars := mux.Vars(r)
	assetName := vars["asset_name"]

	director := func(req *http.Request) {
		req.Header = coreHttp.RemoveSkygearHeader(req.Header)

		query := req.URL.Query()
		pipeline = query.Get(QueryNamePipeline)
		_, hasPipeline = query[QueryNamePipeline]
		// Do not support range request if image processing query is present.
		if hasPipeline {
			req.Header.Del("Range")
			req.Header.Del("If-Range")
		}

		// NOTE(louis): The err is ignored here because we have no way to return it.
		// However, this function does not return error normally.
		// The known condition that err could be returned is fail to sign
		// which is a configuration problem.
		u, signed, _ := h.CloudStorageProvider.RewriteGetURL(req.URL, assetName)
		originallySigned = signed

		req.URL = u

		// Override the Host header
		req.Host = ""
		req.Header.Set("Host", u.Hostname())
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
		if !valid || !hasPipeline {
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
