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
			rootRouter := NewRouter()

			r, _ := http.NewRequest("GET", "/healthz", nil)
			w := httptest.NewRecorder()

			rootRouter.ServeHTTP(w, r)

			So(w.Body.Bytes(), ShouldResemble, []byte("OK"))
		})
	})
}
