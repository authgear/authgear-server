package webapp

import (
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/auth/config"
)

func TestCSPMiddleware(t *testing.T) {
	Convey("CSPMiddleware", t, func() {
		dummy := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
		middleware := &CSPMiddleware{}
		h := middleware.Handle(dummy)

		Convey("no clients", func() {
			middleware.Config = &config.OAuthConfig{}
			w := httptest.NewRecorder()
			r, _ := http.NewRequest("GET", "/", nil)
			h.ServeHTTP(w, r)

			So(w.Result().Header.Get("Content-Security-Policy"), ShouldEqual, "frame-ancestors 'self';")
		})

		Convey("one client", func() {
			middleware.Config = &config.OAuthConfig{
				Clients: []config.OAuthClientConfig{{
					"redirect_uris": []interface{}{
						"https://example.com/path?q=1",
					},
				}},
			}
			w := httptest.NewRecorder()
			r, _ := http.NewRequest("GET", "/", nil)
			h.ServeHTTP(w, r)

			So(w.Result().Header.Get("Content-Security-Policy"), ShouldEqual, "frame-ancestors https://example.com 'self';")
		})

		Convey("more than one clients", func() {
			middleware.Config = &config.OAuthConfig{
				Clients: []config.OAuthClientConfig{
					{
						"redirect_uris": []interface{}{
							"https://example.com/path?q=1",
						},
					},
					{
						"redirect_uris": []interface{}{
							"https://app.com/path?q=1",
						},
					},
				},
			}
			w := httptest.NewRecorder()
			r, _ := http.NewRequest("GET", "/", nil)
			h.ServeHTTP(w, r)

			So(w.Result().Header.Get("Content-Security-Policy"), ShouldEqual, "frame-ancestors https://example.com https://app.com 'self';")
		})

		Convey("include https redirect URIs", func() {
			middleware.Config = &config.OAuthConfig{
				Clients: []config.OAuthClientConfig{{
					"redirect_uris": []interface{}{
						"https://example.com/path?q=1",
						"http://example.com/path?q=1",
						"com.example://host/path",
					},
				}},
			}
			w := httptest.NewRecorder()
			r, _ := http.NewRequest("GET", "/", nil)
			h.ServeHTTP(w, r)

			So(w.Result().Header.Get("Content-Security-Policy"), ShouldEqual, "frame-ancestors https://example.com 'self';")
		})

		Convey("include http redirect URIs if host is localhost", func() {
			middleware.Config = &config.OAuthConfig{
				Clients: []config.OAuthClientConfig{{
					"redirect_uris": []interface{}{
						"http://127.0.0.1/path?q=1",
						"http://127.0.0.1:8080/path?q=1",
						"http://[::1]/path?q=1",
						"http://[::1]:8080/path?q=1",
						"http://localhost/path?q=1",
						"http://localhost:8080/path?q=1",
						"http://foo.localhost/path?q=1",
						"http://foo.localhost:8080/path?q=1",

						"http://example.com/path?q=1",
						"http://192.168.1.1/path?q=1",
						"http://foolocalhost/path?q=1",
					},
				}},
			}
			w := httptest.NewRecorder()
			r, _ := http.NewRequest("GET", "/", nil)
			h.ServeHTTP(w, r)

			So(w.Result().Header.Get("Content-Security-Policy"), ShouldEqual, "frame-ancestors http://127.0.0.1 http://127.0.0.1:8080 http://[::1] http://[::1]:8080 http://localhost http://localhost:8080 http://foo.localhost http://foo.localhost:8080 'self';")
		})
	})
}
