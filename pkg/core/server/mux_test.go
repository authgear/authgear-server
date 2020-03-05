package server

import (
	"net/http"
	"net/http/httptest"
	"testing"

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
				GearPathPrefix: "/_mygear",
			})

			r, _ := http.NewRequest("GET", "/healthz", nil)
			w := httptest.NewRecorder()

			rootRouter.ServeHTTP(w, r)

			So(w.Body.Bytes(), ShouldResemble, []byte("OK"))
		})

		Convey("path prefix", func() {
			rootRouter, appRouter := NewRouterWithOption(Option{
				GearPathPrefix: "/_mygear",
			})

			appRouter.NewRoute().Path("/foobar").Handler(FactoryToHandler(&HandlerFactory{})).Methods("GET")

			r, _ := http.NewRequest("GET", "/_mygear/foobar", nil)
			w := httptest.NewRecorder()

			rootRouter.ServeHTTP(w, r)

			So(w.Body.Bytes(), ShouldResemble, []byte("OK"))
		})
	})
}
