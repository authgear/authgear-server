package handler

import (
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/gorilla/mux"
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
		var err error
		u, err := url.Parse(req.Header.Get(coreHttp.HeaderGearEndpoint))
		if err != nil {
			panic(err)
		}
		req.URL = u
		req.URL.Path = path
		req.URL.RawQuery = query
		req.URL.Fragment = fragment
	}
	return &httputil.ReverseProxy{Director: director}
}

func rewriteHandler(p *httputil.ReverseProxy, restPathIdentifier string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		r.URL.Path = "/" + mux.Vars(r)[restPathIdentifier]
		p.ServeHTTP(w, r)
	}
}
