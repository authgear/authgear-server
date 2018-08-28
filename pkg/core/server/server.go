package server

import (
	"net/http"
	"reflect"
	"time"

	"github.com/gorilla/mux"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/handler"
)

// Server embeds a net/http server and has a gorillax mux internally
type Server struct {
	*http.Server

	router          *mux.Router
	dependencyGraph dependencyGraph
}

type dependencyGraph struct {
	dbProvider *db.DBProvider
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

// SetDBProvider set a DB provider implementation to dependency graph
func (s *Server) SetDBProvider(dbProvider db.DBProvider) {
	s.dependencyGraph.dbProvider = &dbProvider
}

// Handle delegates gorilla mux Handler, and accept a HandlerFactory instead of Handler
func (s *Server) Handle(path string, hf handler.HandlerFactory) *mux.Route {
	return s.router.NewRoute().Path(path).Handler(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		// mock tenant configuration
		configuration := config.TenantConfiguration{
			DBConnectionStr: "public",
		}

		h := hf.NewHandler()

		s.injectDependency(&h, configuration)

		h.Handle(handler.Context{
			ResponseWriter:      rw,
			Request:             r,
			TenantConfiguration: configuration,
		})
	}))
}

func (s Server) injectDependency(h *handler.Handler, configuration config.TenantConfiguration) {
	t := reflect.TypeOf(h).Elem()
	v := reflect.ValueOf(h).Elem()

	numField := t.NumField()
	for i := 0; i < numField; i++ {
		dependencyName := t.Field(i).Tag.Get("dependency")
		field := v.Field(i)
		field.Set(reflect.ValueOf(s.newGet(dependencyName, configuration)))
	}
}

func (s *Server) newGet(dependencyName string, tConfig config.TenantConfiguration) interface{} {
	switch dependencyName {
	case "DB":
		return func() db.IDB {
			dbProvider := s.dependencyGraph.dbProvider
			if dbProvider == nil {
				return nil
			}

			return (*dbProvider).GetDB(tConfig)
		}
	default:
		return nil
	}
}
