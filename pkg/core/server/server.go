package server

import (
	"net/http"
	"time"

	"github.com/skygeario/skygear-server/pkg/core/auth/authz"

	"github.com/gorilla/mux"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/handler"
)

// Server embeds a net/http server and has a gorillax mux internally
type Server struct {
	*http.Server

	router  *mux.Router
	devMode bool
}

// NewServer create a new Server
func NewServer(addr string, devMode bool) Server {
	router := mux.NewRouter()

	srv := &http.Server{
		Addr:         addr,
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
		Handler:      router,
	}

	return Server{
		router:  router,
		Server:  srv,
		devMode: devMode,
	}
}

// Handle delegates gorilla mux Handler, and accept a HandlerFactory instead of Handler
func (s *Server) Handle(path string, hf handler.Factory) *mux.Route {
	// if devMode is true, read configuration from environment variables
	var envConfiguration config.TenantConfiguration
	if s.devMode {
		envConfiguration = config.TenantConfiguration{}
		envConfiguration.ReadFromEnv()
	}

	return s.router.NewRoute().Path(path).Handler(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		var configuration config.TenantConfiguration
		if s.devMode {
			configuration = envConfiguration
		} else {
			configuration = config.GetTenantConfig(r)
		}

		h := hf.NewHandler(r.Context(), configuration)

		ctx := handler.AuthenticationContext{}

		if policy, ok := h.(authz.Policy); ok {
			if err := policy.IsAllowed(r, ctx); err != nil {
				// TODO:
				// handle error properly
				http.Error(rw, err.Error(), http.StatusUnauthorized)
				return
			}
		}

		h.ServeHTTP(rw, r, ctx)
	}))
}
