package server

import (
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/handler"
)

// Server embeds a net/http server and has a gorillax mux internally
type Server struct {
	*http.Server

	router *mux.Router
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

// Handle delegates gorilla mux Handler, and accept a HandlerFactory instead of Handler
func (s *Server) Handle(path string, hf handler.Factory) *mux.Route {
	return s.router.NewRoute().Path(path).Handler(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		// mock tenant configuration
		configuration := config.TenantConfiguration{
			DBConnectionStr: "public",
		}

		h := hf.NewHandler(configuration)

		h.Handle(handler.Context{
			ResponseWriter: rw,
			Request:        r,
		})
	}))
}
