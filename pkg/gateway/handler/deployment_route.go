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
	director := func(req *http.Request) {
		originalPath := req.URL.Path
		coreHttp.SetForwardedHeaders(req)

		ctx := model.GatewayContextFromContext(req.Context())

		forwardURL, err := getForwardURL(req.URL, ctx.RouteMatch)
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
		coreHttp.FixupCORSHeaders(rw, resp)
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
		// query & fragment are not passed to backend
		break
	default:
		panic("unexpected deployment route type")
	}

	return forwardURL, nil
}
