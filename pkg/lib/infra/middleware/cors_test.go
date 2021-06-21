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

	fixture := func(method string, origin string, specs []string) (r *http.Request, h http.Handler) {
		r, _ = http.NewRequest(method, "", nil)
		if origin != "" {
			r.Header.Set("Origin", origin)
		}

		m := CORSMiddleware{
			Config: &config.HTTPConfig{
				AllowedOrigins: specs,
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
			req, handler := fixture("OPTIONS", "http://test.example.com", []string{"example.*"})
			resp := httptest.NewRecorder()
			handler.ServeHTTP(resp, req)

			So(resp.Header().Get("Vary"), ShouldEqual, "Origin")
			So(resp.Header().Get("Access-Control-Allow-Origin"), ShouldBeEmpty)
			So(resp.Header().Get("Access-Control-Allow-Credentials"), ShouldBeEmpty)
		})

		Convey("should handle OPTIONS request", func() {
			req, handler := fixture("OPTIONS", "http://test.example.com", []string{"*.example.com"})
			resp := httptest.NewRecorder()
			handler.ServeHTTP(resp, req)

			So(resp.Header().Get("Vary"), ShouldEqual, "Origin")
			So(resp.Header().Get("Access-Control-Allow-Origin"), ShouldEqual, "http://test.example.com")
			So(resp.Header().Get("Access-Control-Allow-Credentials"), ShouldEqual, "true")
			So(resp.Body.Len(), ShouldEqual, 0)
		})

		Convey("should handle localhost", func() {
			req, handler := fixture("OPTIONS", "http://localhost:3000", []string{"localhost:3000"})
			resp := httptest.NewRecorder()
			handler.ServeHTTP(resp, req)

			So(resp.Header().Get("Vary"), ShouldEqual, "Origin")
			So(resp.Header().Get("Access-Control-Allow-Origin"), ShouldEqual, "http://localhost:3000")
			So(resp.Header().Get("Access-Control-Allow-Credentials"), ShouldEqual, "true")
			So(resp.Body.Len(), ShouldEqual, 0)
		})

		Convey("should handle POST request", func() {
			req, handler := fixture("POST", "http://test.example.com", []string{"*.example.com"})
			resp := httptest.NewRecorder()
			handler.ServeHTTP(resp, req)

			So(resp.Header().Get("Vary"), ShouldEqual, "Origin")
			So(resp.Header().Get("Access-Control-Allow-Origin"), ShouldEqual, "http://test.example.com")
			So(resp.Header().Get("Access-Control-Allow-Credentials"), ShouldEqual, "true")
			So(resp.Body.Bytes(), ShouldResemble, testBody)
		})

		Convey("should handle request with request methods/headers", func() {
			req, handler := fixture("OPTIONS", "http://test.example.com", []string{"*.example.com"})
			req.Header.Set("Access-Control-Request-Method", "GET")
			req.Header.Set("Access-Control-Request-Headers", "Cookie")
			resp := httptest.NewRecorder()
			handler.ServeHTTP(resp, req)

			So(resp.Header().Get("Access-Control-Allow-Methods"), ShouldEqual, "GET")
			So(resp.Header().Get("Access-Control-Allow-Headers"), ShouldEqual, "Cookie")
		})

		Convey("should echo request origin as allowed origin", func() {
			req, handler := fixture("OPTIONS", "https://example.com", []string{"*"})
			resp := httptest.NewRecorder()
			handler.ServeHTTP(resp, req)

			So(resp.Header().Get("Vary"), ShouldEqual, "Origin")
			So(resp.Header().Get("Access-Control-Allow-Origin"), ShouldEqual, "https://example.com")
			So(resp.Header().Get("Access-Control-Allow-Credentials"), ShouldEqual, "true")
		})

		Convey("should not handle request with not allowed origin", func() {
			req, handler := fixture("OPTIONS", "http://example1.com", []string{"*.example.com"})
			resp := httptest.NewRecorder()
			handler.ServeHTTP(resp, req)

			So(resp.Header().Get("Vary"), ShouldEqual, "Origin")
			So(resp.Header().Get("Access-Control-Allow-Origin"), ShouldBeEmpty)
			So(resp.Header().Get("Access-Control-Allow-Credentials"), ShouldBeEmpty)
		})
	})
}
