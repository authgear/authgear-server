package web

import (
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestStaticSecurityHeadersMiddleware(t *testing.T) {
	Convey("StaticSecurityHeadersMiddleware", t, func() {
		middleware := &StaticSecurityHeadersMiddleware{}
		makeHandler := func() http.Handler {
			dummy := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
			h := middleware.Handle(dummy)
			return h
		}

		Convey("disable content type sniffing", func() {
			w := httptest.NewRecorder()
			r, _ := http.NewRequest("GET", "/", nil)
			makeHandler().ServeHTTP(w, r)

			So(w.Result().Header.Get("X-Content-Type-Options"), ShouldEqual, "nosniff")
		})
	})
}
