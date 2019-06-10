package middleware

import (
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/skygeario/skygear-server/pkg/gateway/db"
	gatewayModel "github.com/skygeario/skygear-server/pkg/gateway/model"
)

type FindCloudCodeMiddleware struct {
	RestPathIdentifier string
	Store              db.GatewayStore
}

func (f FindCloudCodeMiddleware) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := gatewayModel.GatewayContextFromContext(r.Context())
		app := ctx.App

		path := "/" + mux.Vars(r)[f.RestPathIdentifier]
		routes, err := f.Store.GetLastDeploymentRoutes(app)
		if err != nil {
			http.Error(w, "Fail to get cloud code routes", http.StatusBadRequest)
			return
		}

		matchedRoute := findMatchedRoute(path, routes)
		if matchedRoute == nil {
			http.Error(w, "Fail to match cloud code route", http.StatusBadRequest)
			return
		}
		ctx.DeploymentRoute = *matchedRoute
		r = r.WithContext(gatewayModel.ContextWithGatewayContext(r.Context(), ctx))

		next.ServeHTTP(w, r)
	})
}

func findMatchedRoute(reqPath string, routes []*gatewayModel.DeploymentRoute) *gatewayModel.DeploymentRoute {
	// convert path to /path/ format
	testReqPath := appendtrailingSlash(reqPath)

	var matchedRoute *gatewayModel.DeploymentRoute
	matchedLen := 0
	for _, r := range routes {
		if reqPath == r.Path {
			return r
		}
		testPath := appendtrailingSlash(r.Path)
		if !strings.HasPrefix(testReqPath, testPath) {
			continue
		}
		if len(r.Path) > matchedLen {
			matchedLen = len(r.Path)
			matchedRoute = r
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
