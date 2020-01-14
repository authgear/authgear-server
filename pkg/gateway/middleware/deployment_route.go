package middleware

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/skygeario/skygear-server/pkg/gateway/model"
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

		match := model.MatchRoute(path, app.Config.DeploymentRoutes)
		if match == nil {
			http.Error(w, "Fail to match deployment route", http.StatusNotFound)
			return
		}
		ctx.RouteMatch = *match
		r = r.WithContext(gatewayModel.ContextWithGatewayContext(r.Context(), ctx))

		next.ServeHTTP(w, r)
	})
}
