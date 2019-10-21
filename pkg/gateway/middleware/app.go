package middleware

import (
	"net/http"

	"github.com/skygeario/skygear-server/pkg/core/logging"

	"github.com/skygeario/skygear-server/pkg/core/config"
	gatewayModel "github.com/skygeario/skygear-server/pkg/gateway/model"
	"github.com/skygeario/skygear-server/pkg/gateway/store"
)

type FindAppMiddleware struct {
	Store store.GatewayStore
}

func (f FindAppMiddleware) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger := logging.NewFactoryFromRequest(r, logging.NewDefaultMaskedTextFormatter(nil)).NewLogger("app-finder")

		host := r.Host
		app := gatewayModel.App{}
		if err := f.Store.GetAppByDomain(host, &app); err != nil {
			logger.WithError(err).Error("failed to found app")
			http.Error(w, "Fail to found app", http.StatusBadRequest)
			return
		}

		routes, err := f.Store.GetLastDeploymentRoutes(app)
		if err != nil {
			logger.WithError(err).Error("failed to get deployment routes")
			http.Error(w, "Fail to get deployment routes", http.StatusInternalServerError)
			return
		}

		for _, route := range routes {
			app.Config.DeploymentRoutes = append(app.Config.DeploymentRoutes, config.DeploymentRoute{
				Version:    route.Version,
				Path:       route.Path,
				Type:       string(route.Type),
				TypeConfig: route.TypeConfig,
			})
		}

		hooks, err := f.Store.GetLastDeploymentHooks(app)
		if store.IsNotFound(err) {
			// no hook exists: ignore error
		} else if err != nil {
			logger.WithError(err).Error("failed to get deployment hooks")
			http.Error(w, "Fail to get deployment hooks", http.StatusInternalServerError)
			return
		} else {
			for _, hook := range hooks.Hooks {
				app.Config.Hooks = append(app.Config.Hooks, config.Hook{
					Event: hook.Event,
					URL:   hook.URL,
				})
			}
		}

		ctx := gatewayModel.GatewayContextFromContext(r.Context())
		ctx.App = app

		r = r.WithContext(gatewayModel.ContextWithGatewayContext(r.Context(), ctx))

		next.ServeHTTP(w, r)
	})
}
