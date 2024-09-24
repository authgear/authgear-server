package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/log"
)

var testBody = []byte{1, 2, 3}

func TestCORSMiddleware(t *testing.T) {
	type FixtureConfig struct {
		Method string
		URL    string
		Origin string
		Specs  []string
		Env    string
	}

	fixture := func(cfg FixtureConfig) (r *http.Request, h http.Handler) {
		r, _ = http.NewRequest(cfg.Method, cfg.URL, nil)
		if cfg.Origin != "" {
			r.Header.Set("Origin", cfg.Origin)
		}

		m := CORSMiddleware{
			Matcher: &CORSMatcher{
				Config: &config.HTTPConfig{
					AllowedOrigins: cfg.Specs,
				},
				OAuthConfig: &config.OAuthConfig{
					Clients: []config.OAuthClientConfig{
						{
							RedirectURIs: []string{
								"http://myapp.example.com/redrect",
							},
							PreAuthenticatedURLAllowedOrigins: []string{
								"http://preauthenticatedurl.example.com",
							},
						},
					},
				},
				SAMLConfig: &config.SAMLConfig{
					ServiceProviders: []*config.SAMLServiceProviderConfig{
						{
							AcsURLs:        []string{"http://acs.example.com"},
							SLOCallbackURL: "http://slo.example.com",
						},
					},
				},
				CORSAllowedOrigins: config.CORSAllowedOrigins(cfg.Env),
			},
			Logger: CORSMiddlewareLogger{log.Null},
		}
		h = m.Handle(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write(testBody)
		}))

		return
	}

	Convey("Test CORSMiddleware", t, func() {
		Convey("should not handle request when CORS config is invalid", func() {
			req, handler := fixture(FixtureConfig{
				Method: "OPTIONS",
				URL:    "http://test.example.com",
				Origin: "http://test.example.com",
				Specs:  []string{"example.*"},
				Env:    "",
			})

			resp := httptest.NewRecorder()
			handler.ServeHTTP(resp, req)

			So(resp.Header().Get("Vary"), ShouldEqual, "Origin")
			So(resp.Header().Get("Access-Control-Allow-Origin"), ShouldBeEmpty)
			So(resp.Header().Get("Access-Control-Allow-Credentials"), ShouldBeEmpty)
		})

		Convey("should handle OPTIONS request", func() {
			req, handler := fixture(FixtureConfig{
				Method: "OPTIONS",
				URL:    "http://www.example.com",
				Origin: "http://test.example.com",
				Specs:  []string{"*.example.com"},
				Env:    "",
			})

			resp := httptest.NewRecorder()
			handler.ServeHTTP(resp, req)

			So(resp.Header().Get("Vary"), ShouldEqual, "Origin")
			So(resp.Header().Get("Access-Control-Allow-Origin"), ShouldEqual, "http://test.example.com")
			So(resp.Header().Get("Access-Control-Allow-Credentials"), ShouldEqual, "true")
			So(resp.Body.Len(), ShouldEqual, 0)
		})

		Convey("should always allow host", func() {
			req, handler := fixture(FixtureConfig{
				Method: "OPTIONS",
				URL:    "http://www.example.com:3000",
				Origin: "http://www.example.com:3000",
				Specs:  nil,
				Env:    "",
			})

			resp := httptest.NewRecorder()
			handler.ServeHTTP(resp, req)

			So(resp.Header().Get("Vary"), ShouldEqual, "Origin")
			So(resp.Header().Get("Access-Control-Allow-Origin"), ShouldEqual, "http://www.example.com:3000")
			So(resp.Header().Get("Access-Control-Allow-Credentials"), ShouldEqual, "true")
			So(resp.Body.Len(), ShouldEqual, 0)
		})

		Convey("should handle localhost", func() {
			req, handler := fixture(FixtureConfig{
				Method: "OPTIONS",
				URL:    "http://localhost:3000",
				Origin: "http://localhost:3000",
				Specs:  []string{"localhost:3000"},
				Env:    "",
			})

			resp := httptest.NewRecorder()
			handler.ServeHTTP(resp, req)

			So(resp.Header().Get("Vary"), ShouldEqual, "Origin")
			So(resp.Header().Get("Access-Control-Allow-Origin"), ShouldEqual, "http://localhost:3000")
			So(resp.Header().Get("Access-Control-Allow-Credentials"), ShouldEqual, "true")
			So(resp.Body.Len(), ShouldEqual, 0)
		})

		Convey("should handle POST request", func() {
			req, handler := fixture(FixtureConfig{
				Method: "POST",
				URL:    "http://www.example.com",
				Origin: "http://test.example.com",
				Specs:  []string{"*.example.com"},
				Env:    "",
			})

			resp := httptest.NewRecorder()
			handler.ServeHTTP(resp, req)

			So(resp.Header().Get("Vary"), ShouldEqual, "Origin")
			So(resp.Header().Get("Access-Control-Allow-Origin"), ShouldEqual, "http://test.example.com")
			So(resp.Header().Get("Access-Control-Allow-Credentials"), ShouldEqual, "true")
			So(resp.Body.Bytes(), ShouldResemble, testBody)
		})

		Convey("should handle request with request methods/headers", func() {
			req, handler := fixture(FixtureConfig{
				Method: "OPTIONS",
				URL:    "http://www.example.com",
				Origin: "http://test.example.com",
				Specs:  []string{"*.example.com"},
				Env:    "",
			})

			req.Header.Set("Access-Control-Request-Method", "GET")
			req.Header.Set("Access-Control-Request-Headers", "Cookie")
			resp := httptest.NewRecorder()
			handler.ServeHTTP(resp, req)

			So(resp.Header().Get("Access-Control-Allow-Methods"), ShouldEqual, "GET")
			So(resp.Header().Get("Access-Control-Allow-Headers"), ShouldEqual, "Cookie")
		})

		Convey("should echo request origin as allowed origin", func() {
			req, handler := fixture(FixtureConfig{
				Method: "OPTIONS",
				URL:    "https://example.com",
				Origin: "https://example.com",
				Specs:  []string{"*"},
				Env:    "",
			})

			resp := httptest.NewRecorder()
			handler.ServeHTTP(resp, req)

			So(resp.Header().Get("Vary"), ShouldEqual, "Origin")
			So(resp.Header().Get("Access-Control-Allow-Origin"), ShouldEqual, "https://example.com")
			So(resp.Header().Get("Access-Control-Allow-Credentials"), ShouldEqual, "true")
		})

		Convey("should not handle request with not allowed origin", func() {
			req, handler := fixture(FixtureConfig{
				Method: "OPTIONS",
				URL:    "http://www.example.com",
				Origin: "http://example1.com",
				Specs:  []string{"*.example.com"},
				Env:    "",
			})

			resp := httptest.NewRecorder()
			handler.ServeHTTP(resp, req)

			So(resp.Header().Get("Vary"), ShouldEqual, "Origin")
			So(resp.Header().Get("Access-Control-Allow-Origin"), ShouldBeEmpty)
			So(resp.Header().Get("Access-Control-Allow-Credentials"), ShouldBeEmpty)
		})

		Convey("should allow origin through environment variable config", func() {
			req, handler := fixture(FixtureConfig{
				Method: "POST",
				URL:    "http://www.example.com",
				Origin: "http://test.example.com",
				Specs:  []string{""},
				Env:    "test.example.com,test2.example.com",
			})

			resp := httptest.NewRecorder()
			handler.ServeHTTP(resp, req)

			So(resp.Header().Get("Vary"), ShouldEqual, "Origin")
			So(resp.Header().Get("Access-Control-Allow-Origin"), ShouldEqual, "http://test.example.com")
			So(resp.Header().Get("Access-Control-Allow-Credentials"), ShouldEqual, "true")
			So(resp.Body.Bytes(), ShouldResemble, testBody)
		})

		Convey("should allow origin in oauth client redirect uris", func() {
			req, handler := fixture(FixtureConfig{
				Method: "POST",
				URL:    "http://www.example.com",
				Origin: "http://myapp.example.com",
				Specs:  []string{""},
				Env:    "",
			})

			resp := httptest.NewRecorder()
			handler.ServeHTTP(resp, req)

			So(resp.Header().Get("Vary"), ShouldEqual, "Origin")
			So(resp.Header().Get("Access-Control-Allow-Origin"), ShouldEqual, "http://myapp.example.com")
			So(resp.Header().Get("Access-Control-Allow-Credentials"), ShouldEqual, "true")
			So(resp.Body.Bytes(), ShouldResemble, testBody)
		})

		Convey("should allow origin in oauth client x_pre_authenticated_url_allowed_origins", func() {
			req, handler := fixture(FixtureConfig{
				Method: "POST",
				URL:    "http://www.example.com",
				Origin: "http://preauthenticatedurl.example.com",
				Specs:  []string{""},
				Env:    "",
			})

			resp := httptest.NewRecorder()
			handler.ServeHTTP(resp, req)

			So(resp.Header().Get("Vary"), ShouldEqual, "Origin")
			So(resp.Header().Get("Access-Control-Allow-Origin"), ShouldEqual, "http://preauthenticatedurl.example.com")
			So(resp.Header().Get("Access-Control-Allow-Credentials"), ShouldEqual, "true")
			So(resp.Body.Bytes(), ShouldResemble, testBody)
		})

		Convey("should allow origin in saml sp acs_urls", func() {
			req, handler := fixture(FixtureConfig{
				Method: "POST",
				URL:    "http://www.example.com",
				Origin: "http://acs.example.com",
				Specs:  []string{""},
				Env:    "",
			})

			resp := httptest.NewRecorder()
			handler.ServeHTTP(resp, req)

			So(resp.Header().Get("Vary"), ShouldEqual, "Origin")
			So(resp.Header().Get("Access-Control-Allow-Origin"), ShouldEqual, "http://acs.example.com")
			So(resp.Header().Get("Access-Control-Allow-Credentials"), ShouldEqual, "true")
			So(resp.Body.Bytes(), ShouldResemble, testBody)
		})

		Convey("should allow origin in saml sp slo_callback_url", func() {
			req, handler := fixture(FixtureConfig{
				Method: "POST",
				URL:    "http://www.example.com",
				Origin: "http://slo.example.com",
				Specs:  []string{""},
				Env:    "",
			})

			resp := httptest.NewRecorder()
			handler.ServeHTTP(resp, req)

			So(resp.Header().Get("Vary"), ShouldEqual, "Origin")
			So(resp.Header().Get("Access-Control-Allow-Origin"), ShouldEqual, "http://slo.example.com")
			So(resp.Header().Get("Access-Control-Allow-Credentials"), ShouldEqual, "true")
			So(resp.Body.Bytes(), ShouldResemble, testBody)
		})
	})
}
