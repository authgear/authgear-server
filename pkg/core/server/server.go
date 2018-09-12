package server

import (
	"net/http"
	"time"

	"github.com/skygeario/skygear-server/pkg/core/auth"
	"github.com/skygeario/skygear-server/pkg/core/auth/authn"
	"github.com/skygeario/skygear-server/pkg/core/auth/authz"

	"github.com/gorilla/mux"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/handler"
)

// Server embeds a net/http server and has a gorillax mux internally
type Server struct {
	*http.Server

	router                  *mux.Router
	authInfoResolverFactory authn.AuthInfoResolverFactory
}

// NewServer create a new Server
func NewServer(addr string) Server {
	router := mux.NewRouter()

	srv := &http.Server{
		Addr:         addr,
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
		Handler:      router,
	}

	return Server{
		router: router,
		Server: srv,
	}
}

func (s *Server) SetAuthInfoResolverFactory(factory authn.AuthInfoResolverFactory) {
	s.authInfoResolverFactory = factory
}

// Handle delegates gorilla mux Handler, and accept a HandlerFactory instead of Handler
func (s *Server) Handle(path string, hf handler.Factory) *mux.Route {
	return s.router.NewRoute().Path(path).Handler(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		configuration := config.GetTenantConfig(r)

		h := hf.NewHandler(r.Context(), configuration)

		var authInfo auth.AuthInfo
		if s.authInfoResolverFactory != nil {
			resolver := s.authInfoResolverFactory.NewResolver(r.Context(), configuration)
			authInfo, _ = resolver.Resolve(r)
		}

		if policyProvider, ok := h.(authz.PolicyProvider); ok {
			policy := policyProvider.ProvideAuthzPolicy()
			if err := policy.IsAllowed(r, authInfo); err != nil {
				// TODO:
				// handle error properly
				http.Error(rw, err.Error(), http.StatusUnauthorized)
				return
			}
		}

		h.ServeHTTP(rw, r, authInfo)
	}))
}

// Use set middlewares to underlying router
func (s *Server) Use(mwf ...mux.MiddlewareFunc) {
	s.router.Use(mwf...)
}
