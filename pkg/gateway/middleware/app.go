package middleware

import (
	"net/http"

	"github.com/skygeario/skygear-server/pkg/core/sentry"

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
		loggerFactory := logging.NewFactoryFromRequest(r,
			logging.NewDefaultLogHook(nil),
			sentry.NewLogHookFromContext(r.Context()),
		)
		logger := loggerFactory.NewLogger("app-finder")

		host := r.Host
		app := gatewayModel.App{}
		if err := f.Store.GetAppByDomain(host, &app); err != nil {
			if store.IsNotFound(err) {
				http.Error(w, "Not found", http.StatusNotFound)
			} else {
				logger.WithError(err).Error("failed to find app")
				http.Error(w, "Internal server error", http.StatusInternalServerError)
			}
			return
		}

		routes, err := f.Store.GetLastDeploymentRoutes(app)
		if err != nil {
			logger.WithError(err).Error("failed to get deployment routes")
			http.Error(w, "Internal server error", http.StatusInternalServerError)
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
			http.Error(w, "Internal server error", http.StatusInternalServerError)
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
