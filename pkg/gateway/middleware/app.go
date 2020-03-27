package middleware

import (
	"net/http"
	"strings"

	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/logging"
	"github.com/skygeario/skygear-server/pkg/core/sentry"

	"github.com/skygeario/skygear-server/pkg/gateway/model"
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
		domain, err := f.getDomain(host)
		if err != nil {
			if store.IsNotFound(err) {
				http.Error(w, "Not found", http.StatusNotFound)
			} else {
				logger.WithError(err).Error("failed to find domain")
				http.Error(w, "Internal server error", http.StatusInternalServerError)
			}
			return
		}

		app, err := f.Store.GetApp(domain.AppID)
		if err != nil {
			if store.IsNotFound(err) {
				http.Error(w, "Not found", http.StatusNotFound)
			} else {
				logger.WithError(err).Error("failed to find app")
				http.Error(w, "Internal server error", http.StatusInternalServerError)
			}
			return
		}

		routes, err := f.Store.GetLastDeploymentRoutes(*app)
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

		hooks, err := f.Store.GetLastDeploymentHooks(*app)
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

		ctx := model.GatewayContextFromContext(r.Context())
		ctx.App = *app
		ctx.Gear = getGearToRoute(domain, r)
		r = r.WithContext(model.ContextWithGatewayContext(r.Context(), ctx))

		next.ServeHTTP(w, r)
	})
}

func (f FindAppMiddleware) getDomain(host string) (*model.Domain, error) {
	domain, err := f.Store.GetDomain(host)
	if err == nil {
		return domain, nil
	}
	if !store.IsNotFound(err) {
		return nil, err
	}

	// try get default domain
	parts := strings.Split(host, ".")
	if len(parts) <= 1 {
		return nil, err
	}
	defaultDomain := strings.Join(parts[1:], ".")
	domain, err = f.Store.GetDefaultDomain(defaultDomain)
	return domain, err
}

func getGearToRoute(domain *model.Domain, r *http.Request) model.Gear {
	if domain.Assignment == model.AssignmentTypeDefault {
		host := r.Host
		if host == domain.Domain {
			// microservices
			// fallback route to gear if necessary
			return model.GetGearByPath(r.URL.Path)
		}
		// get gear from host
		parts := strings.Split(host, ".")
		return model.GetGear(parts[0])
	}
	if domain.Assignment == model.AssignmentTypeMicroservices {
		// fallback route to gear by path
		// return empty string if it is not matched
		return model.GetGearByPath(r.URL.Path)
	}
	return model.Gear(domain.Assignment)
}
