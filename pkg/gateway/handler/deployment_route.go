package handler

import (
	"net/http"
	"net/http/httputil"
	"net/url"

	coreConfig "github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/errors"
	coreHttp "github.com/skygeario/skygear-server/pkg/core/http"
	"github.com/skygeario/skygear-server/pkg/gateway/model"
)

func NewDeploymentRouteHandler() http.HandlerFunc {
	return http.HandlerFunc(handleDeploymentRoute)
}

func handleDeploymentRoute(rw http.ResponseWriter, r *http.Request) {
	ctx := model.GatewayContextFromContext(r.Context())
	routeMatch := model.MatchRoute(r.URL.Path, ctx.App.Config.DeploymentRoutes)
	if routeMatch == nil {
		http.Error(rw, "Not found", http.StatusNotFound)
		return
	}

	director := func(req *http.Request) {
		originalPath := req.URL.Path
		coreHttp.SetForwardedHeaders(req)

		forwardURL, err := getForwardURL(req.URL, *routeMatch)
		if err != nil {
			panic(err)
		}

		req.URL = forwardURL
		// Inject the original path so that
		// downstream can reconstruct the original URL.
		// It does not take backendURL into account.
		req.Header.Add(coreHttp.HeaderHTTPPath, originalPath)
		// Remove tenant config from header.
		coreConfig.WriteTenantConfig(req, nil)
	}
	modifyResponse := func(resp *http.Response) error {
		if routeMatch.Route.Type == model.DeploymentRouteTypeStatic {
			// For static deployment route, we want to pass through the
			// response from backing storage without modification:
			// delete all existing response header.
			headers := rw.Header()
			for name := range headers {
				delete(headers, name)
			}
		} else {
			coreHttp.FixupCORSHeaders(rw, resp)
		}
		if routeMatch.StatusCode != 0 {
			resp.StatusCode = routeMatch.StatusCode
		}
		return nil
	}

	proxy := &httputil.ReverseProxy{
		Director:       director,
		ModifyResponse: modifyResponse,
		ErrorHandler:   reverseProxyErrorHandler,
	}
	proxy.ServeHTTP(rw, r)
}

func getForwardURL(reqURL *url.URL, match model.RouteMatch) (*url.URL, error) {
	var forwardURL *url.URL
	typeConfig := model.RouteTypeConfig(match.Route.TypeConfig)
	switch match.Route.Type {
	case model.DeploymentRouteTypeHTTPService:
		backendURL, err := url.Parse(typeConfig.BackendURL())
		if err != nil {
			return nil, errors.Newf("failed to parse backend URL: %w", err)
		}
		forwardURL = match.ToURL(backendURL)
		forwardURL.RawQuery = reqURL.RawQuery
		forwardURL.Fragment = reqURL.Fragment
		break
	case model.DeploymentRouteTypeStatic:
		backendURL, err := url.Parse(typeConfig.BackendURL())
		if err != nil {
			return nil, errors.Newf("failed to parse backend URL: %w", err)
		}
		forwardURL = match.ToURL(backendURL)
		// Pass query & fragment to backend (asset gear), so image pipeline can be used.
		forwardURL.RawQuery = reqURL.RawQuery
		forwardURL.Fragment = reqURL.Fragment
		break
	default:
		panic("unexpected deployment route type")
	}

	return forwardURL, nil
}
