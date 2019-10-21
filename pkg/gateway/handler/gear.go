package handler

import (
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/gorilla/mux"
	"github.com/skygeario/skygear-server/pkg/core/errors"
	coreHttp "github.com/skygeario/skygear-server/pkg/core/http"
)

// NewGearHandler takes an incoming request and sends it to coresponding
// gear server
func NewGearHandler(restPathIdentifier string) http.HandlerFunc {
	proxy := newGearReverseProxy()
	return rewriteHandler(proxy, restPathIdentifier)
}

func newGearReverseProxy() *httputil.ReverseProxy {
	director := func(req *http.Request) {
		path := req.URL.Path
		query := req.URL.RawQuery
		fragment := req.URL.Fragment
		coreHttp.SetForwardedHeaders(req)

		var err error
		u, err := url.Parse(req.Header.Get(coreHttp.HeaderGearEndpoint))
		if err != nil {
			panic(errors.Newf("failed to parse gear endpoint:%w", err))
		}
		req.URL = u
		req.URL.Path = path
		req.URL.RawQuery = query
		req.URL.Fragment = fragment
	}
	modifyResponse := func(r *http.Response) error {
		// Remove CORS headers because they are managed by this gateway.
		// Auth gear in standalone mode mounts CORS middleware.
		// If we do not remove CORS headers, then the headers will duplicate.
		r.Header.Del("Access-Control-Allow-Origin")
		r.Header.Del("Access-Control-Allow-Credentials")
		r.Header.Del("Access-Control-Allow-Methods")
		r.Header.Del("Access-Control-Allow-Headers")
		return nil
	}

	return &httputil.ReverseProxy{Director: director, ModifyResponse: modifyResponse}
}

func rewriteHandler(p *httputil.ReverseProxy, restPathIdentifier string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		r.URL.Path = "/" + mux.Vars(r)[restPathIdentifier]
		p.ServeHTTP(w, r)
	}
}
