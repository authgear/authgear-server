package server

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"

	. "github.com/smartystreets/goconvey/convey"
)

func TestHealthCheckHandler(t *testing.T) {
	Convey("HealthCheckHandler", t, func() {
		w := httptest.NewRecorder()
		h := http.HandlerFunc(HealthCheckHandler)
		r, _ := http.NewRequest("GET", "", nil)

		h.ServeHTTP(w, r)
		So(w.Result().StatusCode, ShouldEqual, 200)
		So(w.Result().Header.Get("Content-Type"), ShouldEqual, "text/plain")
		So(w.Result().Header.Get("Content-Length"), ShouldEqual, "2")
		So(w.Body.Bytes(), ShouldResemble, []byte("OK"))
	})
}

type HandlerFactory struct{}

func (f *HandlerFactory) NewHandler(r *http.Request) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	})
}

func TestServer(t *testing.T) {
	Convey("Server", t, func() {
		Convey("IsAPIVersioned = false", func() {
			s := NewServerWithOption("0.0.0.0:3000", nil, Option{
				RecoverPanic:   true,
				GearPathPrefix: "/_mygear",
			})

			s.Handle("/foobar", &HandlerFactory{}).Methods("GET")

			r, _ := http.NewRequest("GET", "/_mygear/foobar", nil)
			w := httptest.NewRecorder()

			s.ServeHTTP(w, r)

			So(w.Body.Bytes(), ShouldResemble, []byte("OK"))
		})

		Convey("IsAPIVersioned = true", func() {
			var apiVersion string

			s := NewServerWithOption("0.0.0.0:3000", nil, Option{
				RecoverPanic:   true,
				GearPathPrefix: "/_mygear",
				IsAPIVersioned: true,
			})

			s.Use(func(next http.Handler) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					apiVersion = mux.Vars(r)["api_version"]
					next.ServeHTTP(w, r)
				})
			})

			s.Handle("/foobar", &HandlerFactory{}).Methods("GET")

			r, _ := http.NewRequest("GET", "/_mygear/v2.0/foobar", nil)
			w := httptest.NewRecorder()

			s.ServeHTTP(w, r)

			So(w.Body.Bytes(), ShouldResemble, []byte("OK"))
			So(apiVersion, ShouldEqual, "v2.0")
		})
	})
}
