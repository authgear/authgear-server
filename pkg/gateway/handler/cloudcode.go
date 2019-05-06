package handler

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"path"

	coreConfig "github.com/skygeario/skygear-server/pkg/core/config"
	coreHttp "github.com/skygeario/skygear-server/pkg/core/http"
	"github.com/skygeario/skygear-server/pkg/gateway/config"
	"github.com/skygeario/skygear-server/pkg/gateway/model"
)

func NewCloudCodeHandler(routerConfig config.RouterConfig) http.HandlerFunc {
	proxy := newCloudCodeReverseProxy(routerConfig)
	return proxy.ServeHTTP
}

func newCloudCodeReverseProxy(routerConfig config.RouterConfig) *httputil.ReverseProxy {
	director := func(req *http.Request) {
		originalPath := req.URL.Path
		query := req.URL.RawQuery
		fragment := req.URL.Fragment

		ctx := model.GatewayContextFromContext(req.Context())
		cloudCode := ctx.CloudCode

		var err error
		backendURL, err := url.Parse(cloudCode.BackendURL)
		if err != nil {
			panic(err)
		}

		// Handle case that the backend URL does not have trailing slash
		if backendURL.Path == "" {
			backendURL.Path = "/"
		}

		req.URL = backendURL
		req.URL.Path = path.Join(req.URL.Path, cloudCode.TargetPath)
		req.URL.RawQuery = query
		req.URL.Fragment = fragment
		// Inject the original path so that
		// downstream can reconstruct the original URL.
		// It does not take backendURL into account.
		req.Header.Add(coreHttp.HeaderHTTPPath, originalPath)
		// Remove tenant config from header.
		coreConfig.DelTenantConfig(req)
	}

	return &httputil.ReverseProxy{Director: director}
}
