package server

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	. "github.com/smartystreets/goconvey/convey"
)

type HandlerFactory struct{}

func (f *HandlerFactory) NewHandler(r *http.Request) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	})
}

func TestRouter(t *testing.T) {
	Convey("Router", t, func() {
		Convey("/healthz", func() {
			rootRouter, _ := NewRouterWithOption(Option{
				RecoverPanic:   true,
				GearPathPrefix: "/_mygear",
			})

			r, _ := http.NewRequest("GET", "/healthz", nil)
			w := httptest.NewRecorder()

			rootRouter.ServeHTTP(w, r)

			So(w.Body.Bytes(), ShouldResemble, []byte("OK"))
		})

		Convey("IsAPIVersioned = false", func() {
			rootRouter, appRouter := NewRouterWithOption(Option{
				RecoverPanic:   true,
				GearPathPrefix: "/_mygear",
			})

			appRouter.NewRoute().Path("/foobar").Handler(FactoryToHandler(&HandlerFactory{})).Methods("GET")

			r, _ := http.NewRequest("GET", "/_mygear/foobar", nil)
			w := httptest.NewRecorder()

			rootRouter.ServeHTTP(w, r)

			So(w.Body.Bytes(), ShouldResemble, []byte("OK"))
		})

		Convey("IsAPIVersioned = true", func() {
			var apiVersion string

			rootRouter, appRouter := NewRouterWithOption(Option{
				RecoverPanic:   true,
				GearPathPrefix: "/_mygear",
				IsAPIVersioned: true,
			})

			appRouter.Use(func(next http.Handler) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					apiVersion = mux.Vars(r)["api_version"]
					next.ServeHTTP(w, r)
				})
			})

			appRouter.NewRoute().Path("/foobar").Handler(FactoryToHandler(&HandlerFactory{})).Methods("GET")

			r, _ := http.NewRequest("GET", "/_mygear/v2.0/foobar", nil)
			w := httptest.NewRecorder()

			rootRouter.ServeHTTP(w, r)

			So(w.Body.Bytes(), ShouldResemble, []byte("OK"))
			So(apiVersion, ShouldEqual, "v2.0")
		})
	})
}
