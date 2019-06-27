package handler

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"path"
	"strings"

	coreConfig "github.com/skygeario/skygear-server/pkg/core/config"
	coreHttp "github.com/skygeario/skygear-server/pkg/core/http"
	"github.com/skygeario/skygear-server/pkg/gateway/model"
)

func NewDeploymentRouteHandler() http.HandlerFunc {
	proxy := newDeploymentRouteReverseProxy()
	return proxy.ServeHTTP
}

func newDeploymentRouteReverseProxy() *httputil.ReverseProxy {
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

func getForwardURL(reqURL *url.URL, route coreConfig.DeploymentRoute) (*url.URL, error) {
	var forwardURL *url.URL
	var err error
	typeConfig := model.RouteTypeConfig(route.TypeConfig)
	switch route.Type {
	case model.DeploymentRouteTypeFunction, model.DeploymentRouteTypeHTTPHandler:
		forwardURL, err = url.Parse(typeConfig.BackendURL())
		if err != nil {
			return nil, err
		}
		// Handle case that the backend URL does not have trailing slash
		if forwardURL.Path == "" {
			forwardURL.Path = "/"
		}
		forwardURL.Path = path.Join(
			forwardURL.Path,
			typeConfig.TargetPath(),
		)
		break
	case model.DeploymentRouteTypeHTTPService:
		forwardURL, err = url.Parse(typeConfig.BackendURL())
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

		// path.Join will clean the result and the returned path ends in a
		// slash only if it is the root "/".
		// check and add back the trailing slash if necessary
		if trimmedPath != "/" && strings.HasSuffix(trimmedPath, "/") {
			forwardURL.Path = forwardURL.Path + "/"
		}
		break
	default:
		panic("unexpected deployment route type")
	}

	forwardURL.RawQuery = reqURL.RawQuery
	forwardURL.Fragment = reqURL.Fragment

	return forwardURL, nil
}
