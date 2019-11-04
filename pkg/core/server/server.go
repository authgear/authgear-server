package server

import (
	"io"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/skygeario/skygear-server/pkg/core/handler"
	"github.com/skygeario/skygear-server/pkg/core/inject"
	"github.com/skygeario/skygear-server/pkg/core/middleware"
	"github.com/skygeario/skygear-server/pkg/core/sentry"
)

// Server embeds a net/http server and has a gorillax mux internally
type Server struct {
	*http.Server

	router        *mux.Router
	dependencyMap inject.DependencyMap
}

// NewServer create a new Server with default option
func NewServer(
	addr string,
	dependencyMap inject.DependencyMap,
) Server {
	return NewServerWithOption(
		addr,
		dependencyMap,
		DefaultOption(),
	)
}

// NewServerWithOption create a new Server
func NewServerWithOption(
	addr string,
	dependencyMap inject.DependencyMap,
	option Option,
) Server {
	router := mux.NewRouter()
	router.HandleFunc("/healthz", HealthCheckHandler)

	var subRouter *mux.Router
	if option.GearPathPrefix == "" {
		subRouter = router.NewRoute().Subrouter()
	} else {
		subRouter = router.PathPrefix(option.GearPathPrefix).Subrouter()
	}
	srv := Server{
		router: subRouter,
		Server: &http.Server{
			Addr:    addr,
			Handler: router,
		},
		dependencyMap: dependencyMap,
	}

	srv.Use(sentry.Middleware(sentry.DefaultClient.Hub))
	if option.RecoverPanic {
		srv.Use(middleware.RecoverMiddleware{}.Handle)
	}

	return srv
}

// Handle delegates gorilla mux Handler, and accept a HandlerFactory instead of Handler
func (s *Server) Handle(path string, hf handler.Factory) *mux.Route {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := hf.NewHandler(r)
		h.ServeHTTP(w, r)
	})

	return s.router.NewRoute().Path(path).Handler(handler)
}

// Use set middlewares to underlying router
func (s *Server) Use(mwf ...mux.MiddlewareFunc) {
	s.router.Use(mwf...)
}

// HealthCheckHandler is basic handler for server health check
func HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	io.WriteString(w, "OK")
}
