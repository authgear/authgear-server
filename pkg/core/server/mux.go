package server

import (
	"net/http"

	"github.com/gorilla/mux"

	"github.com/skygeario/skygear-server/pkg/core/handler"
	"github.com/skygeario/skygear-server/pkg/core/middleware"
	"github.com/skygeario/skygear-server/pkg/core/sentry"
)

func NewRouterWithOption(option Option) (rootRouter *mux.Router, appRouter *mux.Router) {
	rootRouter = mux.NewRouter()
	rootRouter.HandleFunc("/healthz", HealthCheckHandler)

	if option.GearPathPrefix == "" {
		appRouter = rootRouter.NewRoute().Subrouter()
	} else {
		appRouter = rootRouter.PathPrefix(option.GearPathPrefix).Subrouter()
	}

	appRouter.Use(sentry.Middleware(sentry.DefaultClient.Hub))
	appRouter.Use(middleware.RecoverMiddleware{}.Handle)

	return
}

func FactoryToHandler(f handler.Factory) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := f.NewHandler(r)
		h.ServeHTTP(w, r)
	})
}
