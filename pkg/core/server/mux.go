package server

import (
	"github.com/gorilla/mux"

	"github.com/skygeario/skygear-server/pkg/core/middleware"
	"github.com/skygeario/skygear-server/pkg/core/sentry"
)

func NewRouter() *mux.Router {
	rootRouter := mux.NewRouter()
	rootRouter.Use(sentry.Middleware(sentry.DefaultClient.Hub))
	rootRouter.Use(middleware.RecoverMiddleware{}.Handle)
	return rootRouter
}
