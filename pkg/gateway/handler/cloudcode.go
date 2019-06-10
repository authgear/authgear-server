package handler

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"path"
	"strings"

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

		ctx := model.GatewayContextFromContext(req.Context())
		deploymentRoute := ctx.DeploymentRoute

		forwardURL, err := getForwardURL(req.URL, deploymentRoute)
		if err != nil {
			panic(err)
		}

		req.URL = forwardURL
		// Inject the original path so that
		// downstream can reconstruct the original URL.
		// It does not take backendURL into account.
		req.Header.Add(coreHttp.HeaderHTTPPath, originalPath)
		// Remove tenant config from header.
		coreConfig.DelTenantConfig(req)
	}

	return &httputil.ReverseProxy{Director: director}
}

func getForwardURL(reqURL *url.URL, route model.DeploymentRoute) (*url.URL, error) {
	var forwardURL *url.URL
	var err error
	switch route.Type {
	case model.DeploymentRouteTypeFunction, model.DeploymentRouteTypeHTTPHandler:
		forwardURL, err = url.Parse(route.TypeConfig.BackendURL())
		if err != nil {
			return nil, err
		}
		// Handle case that the backend URL does not have trailing slash
		if forwardURL.Path == "" {
			forwardURL.Path = "/"
		}
		forwardURL.Path = path.Join(
			forwardURL.Path,
			route.TypeConfig.TargetPath(),
		)
		break
	case model.DeploymentRouteTypeHTTPService:
		forwardURL, err = url.Parse(route.TypeConfig.BackendURL())
		if err != nil {
			return nil, err
		}
		// Handle case that the backend URL does not have trailing slash
		if forwardURL.Path == "" {
			forwardURL.Path = "/"
		}
		// remove trailing slash to handle the case that route path has
		// trailing slash but the request path doesn't
		routePath := strings.TrimSuffix(route.Path, "/")
		trimmedPath := strings.TrimPrefix(reqURL.Path, routePath)
		forwardURL.Path = path.Join(
			forwardURL.Path,
			trimmedPath,
		)
		break
	default:
		panic("unexpected deployment route type")
	}

	forwardURL.RawQuery = reqURL.RawQuery
	forwardURL.Fragment = reqURL.Fragment

	return forwardURL, nil
}
