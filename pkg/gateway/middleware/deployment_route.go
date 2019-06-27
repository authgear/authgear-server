package middleware

import (
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/skygeario/skygear-server/pkg/core/config"
	gatewayModel "github.com/skygeario/skygear-server/pkg/gateway/model"
	"github.com/skygeario/skygear-server/pkg/gateway/store"
)

type FindDeploymentRouteMiddleware struct {
	RestPathIdentifier string
	Store              store.GatewayStore
}

func (f FindDeploymentRouteMiddleware) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := gatewayModel.GatewayContextFromContext(r.Context())
		app := ctx.App

		path := "/" + mux.Vars(r)[f.RestPathIdentifier]

		matchedRoute := findMatchedRoute(path, app.Config.DeploymentRoutes)
		if matchedRoute == nil {
			http.Error(w, "Fail to match deployment route", http.StatusNotFound)
			return
		}
		ctx.DeploymentRoute = *matchedRoute
		r = r.WithContext(gatewayModel.ContextWithGatewayContext(r.Context(), ctx))

		next.ServeHTTP(w, r)
	})
}

func findMatchedRoute(reqPath string, routes []config.DeploymentRoute) *config.DeploymentRoute {
	// convert path to /path/ format
	testReqPath := appendtrailingSlash(reqPath)

	var matchedRoute *config.DeploymentRoute
	matchedLen := 0
	for _, r := range routes {
		if reqPath == r.Path {
			return &r
		}
		testPath := appendtrailingSlash(r.Path)
		if !strings.HasPrefix(testReqPath, testPath) {
			continue
		}
		if len(r.Path) > matchedLen {
			matchedLen = len(r.Path)
			// We must copy the route because r is
			// reused in all iterations.
			copied := r
			matchedRoute = &copied
		}
	}
	return matchedRoute
}

func appendtrailingSlash(path string) string {
	if !strings.HasSuffix(path, "/") {
		return path + "/"
	}
	return path
}
