package webapp

import (
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/lib/config"
)

func TestNoProjectCSPMiddleware(t *testing.T) {
	Convey("NoProjectCSPMiddleware", t, func() {
		h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		})

		Convey("should set CSP and allow frame ancestors from env", func() {
			middleware := &NoProjectCSPMiddleware{
				AllowedFrameAncestorsFromEnv: config.AllowedFrameAncestors{"http://example.com"},
			}

			handler := middleware.Handle(h)
			w := httptest.NewRecorder()
			r, _ := http.NewRequest("GET", "/", nil)

			handler.ServeHTTP(w, r)

			So(w.Result().Header.Get("Content-Security-Policy"), ShouldContainSubstring, "frame-ancestors http://example.com")
			So(w.Result().Header.Get("X-Frame-Options"), ShouldEqual, "")
		})

		Convey("should set X-Frame-Options to DENY when no frame ancestors are allowed", func() {
			middleware := &NoProjectCSPMiddleware{
				AllowedFrameAncestorsFromEnv: nil,
			}

			handler := middleware.Handle(h)
			w := httptest.NewRecorder()
			r, _ := http.NewRequest("GET", "/", nil)

			handler.ServeHTTP(w, r)

			So(w.Result().Header.Get("Content-Security-Policy"), ShouldContainSubstring, "frame-ancestors 'none'")
			So(w.Result().Header.Get("X-Frame-Options"), ShouldEqual, "DENY")
		})
	})
}
