package server

import (
	"net/http"

	"github.com/gorilla/mux"

	"github.com/skygeario/skygear-server/pkg/core/handler"
	"github.com/skygeario/skygear-server/pkg/core/middleware"
	"github.com/skygeario/skygear-server/pkg/core/sentry"
)

func NewRouter() *mux.Router {
	rootRouter := mux.NewRouter()
	rootRouter.Use(sentry.Middleware(sentry.DefaultClient.Hub))
	rootRouter.Use(middleware.RecoverMiddleware{}.Handle)
	return rootRouter
}

func FactoryToHandler(f handler.Factory) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := f.NewHandler(r)
		h.ServeHTTP(w, r)
	})
}
