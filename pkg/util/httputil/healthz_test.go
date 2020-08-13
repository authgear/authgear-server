package httputil

import (
	"net/http"
	"net/http/httptest"
	"testing"

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
