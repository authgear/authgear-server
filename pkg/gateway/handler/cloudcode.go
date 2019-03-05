package handler

import (
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/skygeario/skygear-server/pkg/gateway/config"
	"github.com/skygeario/skygear-server/pkg/gateway/model"
)

func NewCloudCodeHandler(routerConfig config.RouterConfig) http.HandlerFunc {
	proxy := newCloudCodeReverseProxy(routerConfig)
	return proxy.ServeHTTP
}

func newCloudCodeReverseProxy(routerConfig config.RouterConfig) *httputil.ReverseProxy {
	director := func(req *http.Request) {
		query := req.URL.RawQuery
		fragment := req.URL.Fragment
		var err error
		u, err := url.Parse(routerConfig.CloudCodeGatewayURL)
		if err != nil {
			panic(err)
		}
		cloudCode := model.CloudCodeFromContext(req.Context())
		req.URL = u
		req.URL.Path = "/function" + cloudCode.TargetPath
		req.URL.RawQuery = query
		req.URL.Fragment = fragment
	}
	return &httputil.ReverseProxy{Director: director}
}
