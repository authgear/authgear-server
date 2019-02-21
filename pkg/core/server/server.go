package server

import (
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/skygeario/skygear-server/pkg/core/auth"
	"github.com/skygeario/skygear-server/pkg/core/auth/authn"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/middleware"
	nextSkyerr "github.com/skygeario/skygear-server/pkg/core/skyerr"
	"github.com/skygeario/skygear-server/pkg/server/skyerr"

	"github.com/gorilla/mux"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/handler"
)

// Server embeds a net/http server and has a gorillax mux internally
type Server struct {
	*http.Server

	router                     *mux.Router
	authContextResolverFactory authn.AuthContextResolverFactory
}

// NewServer create a new Server with default option
func NewServer(
	addr string,
	authContextResolverFactory authn.AuthContextResolverFactory,
) Server {
	return NewServerWithOption(
		addr,
		authContextResolverFactory,
		DefaultOption(),
	)
}

// NewServerWithOption create a new Server
func NewServerWithOption(
	addr string,
	authContextResolverFactory authn.AuthContextResolverFactory,
	option Option,
) Server {
	router := mux.NewRouter()
	router.HandleFunc("/healthz", HealthCheckHandler)

	subRouter := router.NewRoute().Subrouter()
	srv := Server{
		router: subRouter,
		Server: &http.Server{
			Addr:         addr,
			WriteTimeout: time.Second * 15,
			ReadTimeout:  time.Second * 15,
			IdleTimeout:  time.Second * 60,
			Handler:      router,
		},
		authContextResolverFactory: authContextResolverFactory,
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
	return s.router.NewRoute().Path(path).Handler(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		r = auth.InitRequestAuthContext(r)
		r = db.InitRequestDBContext(r)

		configuration := config.GetTenantConfig(r)

		h := hf.NewHandler(r)

		txContext := db.NewTxContextWithContext(r.Context(), configuration)
		resolver := s.authContextResolverFactory.NewResolver(r.Context(), configuration)
		db.WithTx(txContext, func() error {
			return resolver.Resolve(r, auth.NewContextSetterWithContext(r.Context()))
		})

		policy := hf.ProvideAuthzPolicy()
		if err := policy.IsAllowed(r, auth.NewContextGetterWithContext(r.Context())); err != nil {
			// TODO: log
			s.handleAuthzError(rw, err)
			return
		}

		h.ServeHTTP(rw, r)
	}))
}

// Use set middlewares to underlying router
func (s *Server) Use(mwf ...mux.MiddlewareFunc) {
	s.router.Use(mwf...)
}

func (s *Server) handleAuthzError(rw http.ResponseWriter, err error) {
	skyErr := skyerr.MakeError(err)
	httpStatus := nextSkyerr.ErrorDefaultStatusCode(skyErr)
	response := authzErrorResponse{Err: skyErr}
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

type authzErrorResponse struct {
	Err skyerr.Error `json:"error,omitempty"`
}
