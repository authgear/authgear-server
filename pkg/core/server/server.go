package server

import (
	"net/http"
	"time"

	"github.com/skygeario/skygear-server/pkg/core/auth"
	"github.com/skygeario/skygear-server/pkg/core/auth/authn"
	"github.com/skygeario/skygear-server/pkg/core/middleware"

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

	srv := Server{
		router: router,
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

		configuration := config.GetTenantConfig(r)

		h := hf.NewHandler(r)

		resolver := s.authContextResolverFactory.NewResolver(r.Context(), configuration)
		resolver.Resolve(r, auth.NewContextSetterWithContext(r.Context()))

		policy := hf.ProvideAuthzPolicy()
		if err := policy.IsAllowed(r, auth.NewContextGetterWithContext(r.Context())); err != nil {
			// TODO:
			// handle error properly
			http.Error(rw, err.Error(), http.StatusUnauthorized)
			return
		}

		h.ServeHTTP(rw, r)
	}))
}

// Use set middlewares to underlying router
func (s *Server) Use(mwf ...mux.MiddlewareFunc) {
	s.router.Use(mwf...)
}
