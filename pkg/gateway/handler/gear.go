package handler

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/skygeario/skygear-server/pkg/core/errors"
	coreHttp "github.com/skygeario/skygear-server/pkg/core/http"
)

// NewGearHandler takes an incoming request and sends it to coresponding
// gear server
func NewGearHandler() http.Handler {
	return http.HandlerFunc(handleGear)
}

func handleGear(rw http.ResponseWriter, r *http.Request) {
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
	modifyResponse := func(resp *http.Response) error {
		// Remove CORS headers if upstream provides them
		for name := range resp.Header {
			if strings.HasPrefix(name, "Access-Control-") {
				rw.Header().Del(name)
			}
		}
		return nil
	}

	proxy := &httputil.ReverseProxy{Director: director, ModifyResponse: modifyResponse}
	proxy.ServeHTTP(rw, r)
}
