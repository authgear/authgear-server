package webapp

import (
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/skygeario/skygear-server/pkg/core/config"
)

func TestCSPMiddleware(t *testing.T) {
	Convey("CSPMiddleware", t, func() {
		dummy := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
		middleware := &CSPMiddleware{}
		h := middleware.Handle(dummy)

		Convey("no clients", func() {
			w := httptest.NewRecorder()
			r, _ := http.NewRequest("GET", "/", nil)
			h.ServeHTTP(w, r)

			So(w.Result().Header.Get("Content-Security-Policy"), ShouldEqual, "frame-ancestors 'self';")
		})

		Convey("one client", func() {
			middleware.Clients = []config.OAuthClientConfiguration{
				config.OAuthClientConfiguration{
					"redirect_uris": []interface{}{
						"https://example.com/path?q=1",
					},
				},
			}
			w := httptest.NewRecorder()
			r, _ := http.NewRequest("GET", "/", nil)
			h.ServeHTTP(w, r)

			So(w.Result().Header.Get("Content-Security-Policy"), ShouldEqual, "frame-ancestors https://example.com 'self';")
		})

		Convey("more than one clients", func() {
			middleware.Clients = []config.OAuthClientConfiguration{
				config.OAuthClientConfiguration{
					"redirect_uris": []interface{}{
						"https://example.com/path?q=1",
					},
				},
				config.OAuthClientConfiguration{
					"redirect_uris": []interface{}{
						"https://app.com/path?q=1",
					},
				},
			}
			w := httptest.NewRecorder()
			r, _ := http.NewRequest("GET", "/", nil)
			h.ServeHTTP(w, r)

			So(w.Result().Header.Get("Content-Security-Policy"), ShouldEqual, "frame-ancestors https://example.com https://app.com 'self';")
		})

		Convey("include only https? redirect URIs", func() {
			middleware.Clients = []config.OAuthClientConfiguration{
				config.OAuthClientConfiguration{
					"redirect_uris": []interface{}{
						"https://example.com/path?q=1",
						"http://localhost/path?q=1",
						"com.example://host/path",
					},
				},
			}
			w := httptest.NewRecorder()
			r, _ := http.NewRequest("GET", "/", nil)
			h.ServeHTTP(w, r)

			So(w.Result().Header.Get("Content-Security-Policy"), ShouldEqual, "frame-ancestors https://example.com http://localhost 'self';")
		})
	})
}
