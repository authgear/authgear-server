package server

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/skygeario/skygear-server/pkg/core/inject"

	"github.com/skygeario/skygear-server/pkg/core/auth"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/logging"
	"github.com/skygeario/skygear-server/pkg/core/middleware"
	"github.com/skygeario/skygear-server/pkg/core/skyerr"

	"github.com/gorilla/mux"
	"github.com/skygeario/skygear-server/pkg/core/handler"
)

// SetupContextFunc setup context for usage in handler
type SetupContextFunc func(context.Context) context.Context

// CleanupContextFunc cleanup context after handler completed
type CleanupContextFunc func(context.Context)

// Server embeds a net/http server and has a gorillax mux internally
type Server struct {
	*http.Server

	router        *mux.Router
	dependencyMap inject.DependencyMap
	dbPool        db.Pool
	setupCtxFn    SetupContextFunc
	cleanupCtxFn  CleanupContextFunc
}

// NewServer create a new Server with default option
func NewServer(
	addr string,
	dependencyMap inject.DependencyMap,
	dbPool db.Pool,
	setupCtxFn SetupContextFunc,
	cleanupCtxFn CleanupContextFunc,
) Server {
	return NewServerWithOption(
		addr,
		dependencyMap,
		dbPool,
		setupCtxFn,
		cleanupCtxFn,
		DefaultOption(),
	)
}

// NewServerWithOption create a new Server
func NewServerWithOption(
	addr string,
	dependencyMap inject.DependencyMap,
	dbPool db.Pool,
	setupCtxFn SetupContextFunc,
	cleanupCtxFn CleanupContextFunc,
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
			Addr:         addr,
			WriteTimeout: time.Second * 15,
			ReadTimeout:  time.Second * 15,
			IdleTimeout:  time.Second * 60,
			Handler:      router,
		},
		dependencyMap: dependencyMap,
		dbPool:        dbPool,
		setupCtxFn:    setupCtxFn,
		cleanupCtxFn:  cleanupCtxFn,
	}

	if option.RecoverPanic {
		srv.Use(middleware.RecoverMiddleware{
			RecoverHandler: option.RecoverPanicHandler,
		}.Handle)
	}

	return srv
}

// Handle delegates gorilla mux Handler, and accept a HandlerFactory instead of Handler
func (s *Server) Handle(path string, hf handler.Factory) *mux.Route {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// TODO: improve the logger
		log := logging.CreateLoggerWithContext(r.Context(), "server")

		policy := hf.ProvideAuthzPolicy()
		if err := policy.IsAllowed(r, auth.NewContextGetterWithContext(r.Context())); err != nil {
			// TODO: log
			log.WithError(err).Info("authz not allowed")
			s.handleError(w, err)
			return
		}

		h := hf.NewHandler(r)
		h.ServeHTTP(w, r)
	})

	// TODO(middleware): refactor this
	authnMiddleware := middleware.Injecter{
		MiddlewareFactory: middleware.AuthnMiddlewareFactory{},
		Dependency:        s.dependencyMap,
	}.Handle

	h := authnMiddleware(handler)

	return s.router.NewRoute().Path(path).Handler(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		r = auth.InitRequestAuthContext(r)
		r = db.InitRequestDBContext(r, s.dbPool)
		if s.setupCtxFn != nil {
			r = r.WithContext(s.setupCtxFn(r.Context()))
		}
		if s.cleanupCtxFn != nil {
			defer s.cleanupCtxFn(r.Context())
		}

		h.ServeHTTP(rw, r)
	}))
}

// Use set middlewares to underlying router
func (s *Server) Use(mwf ...mux.MiddlewareFunc) {
	s.router.Use(mwf...)
}

func (s *Server) handleError(rw http.ResponseWriter, err error) {
	skyErr := skyerr.MakeError(err)
	httpStatus := skyerr.ErrorDefaultStatusCode(skyErr)
	response := errorResponse{Err: skyErr}
	encoder := json.NewEncoder(rw)
	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(httpStatus)
	encoder.Encode(response)
}

// HealthCheckHandler is basic handler for server health check
func HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	io.WriteString(w, "OK")
}

type errorResponse struct {
	Err skyerr.Error `json:"error,omitempty"`
}
