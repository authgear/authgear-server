package web

import (
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestSecHeadersMiddleware(t *testing.T) {
	Convey("SecHeadersMiddleware", t, func() {
		middleware := &SecHeadersMiddleware{}
		makeHandler := func() http.Handler {
			dummy := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
			h := middleware.Handle(dummy)
			return h
		}

		Convey("disable content type sniffing", func() {
			middleware.FrameAncestors = nil
			w := httptest.NewRecorder()
			r, _ := http.NewRequest("GET", "/", nil)
			makeHandler().ServeHTTP(w, r)

			So(w.Result().Header.Get("X-Content-Type-Options"), ShouldEqual, "nosniff")
		})

		Convey("csp directives", func() {
			middleware.CSPDirectives = []string{
				"default-src 'self'",
				"object-src 'none'",
				"base-uri 'none'",
				"script-src 'self'",
				"block-all-mixed-content",
			}
			w := httptest.NewRecorder()
			r, _ := http.NewRequest("GET", "/", nil)
			makeHandler().ServeHTTP(w, r)

			So(w.Result().Header.Get("Content-Security-Policy"), ShouldEqual, "default-src 'self'; object-src 'none'; base-uri 'none'; script-src 'self'; block-all-mixed-content; frame-ancestors 'none'")
			So(w.Result().Header.Get("X-Frame-Options"), ShouldEqual, "DENY")
		})

		Convey("deny frame embedding", func() {
			middleware.FrameAncestors = nil
			w := httptest.NewRecorder()
			r, _ := http.NewRequest("GET", "/", nil)
			makeHandler().ServeHTTP(w, r)

			So(w.Result().Header.Get("Content-Security-Policy"), ShouldEqual, "frame-ancestors 'none'")
			So(w.Result().Header.Get("X-Frame-Options"), ShouldEqual, "DENY")
		})

		Convey("allow frame embedding from specific URL", func() {
			middleware.FrameAncestors = []string{
				"http://localhost",
			}
			w := httptest.NewRecorder()
			r, _ := http.NewRequest("GET", "/", nil)
			makeHandler().ServeHTTP(w, r)

			So(w.Result().Header.Get("Content-Security-Policy"), ShouldEqual, "frame-ancestors http://localhost")
			So(w.Result().Header.Get("X-Frame-Options"), ShouldEqual, "")
		})
	})
}
