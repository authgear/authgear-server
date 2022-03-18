package web

import (
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestStaticCSPMiddleware(t *testing.T) {
	Convey("StaticCSPMiddleware", t, func() {
		middleware := StaticCSPMiddleware{}
		makeHandler := func() http.Handler {
			dummy := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
			h := middleware.Handle(dummy)
			return h
		}

		Convey("csp directives", func() {
			middleware.CSPDirectives = []string{
				"default-src 'self'",
				"object-src 'none'",
				"base-uri 'none'",
				"script-src 'self'",
				"block-all-mixed-content",
				"frame-ancestors 'none'",
			}
			w := httptest.NewRecorder()
			r, _ := http.NewRequest("GET", "/", nil)
			makeHandler().ServeHTTP(w, r)

			So(w.Result().Header.Get("Content-Security-Policy"), ShouldEqual, "default-src 'self'; object-src 'none'; base-uri 'none'; script-src 'self'; block-all-mixed-content; frame-ancestors 'none'")
		})
	})
}
