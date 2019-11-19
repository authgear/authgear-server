package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/skygeario/skygear-server/pkg/core/config"
	. "github.com/smartystreets/goconvey/convey"
)

var testBody = []byte{1, 2, 3}

func getTestHandler() http.HandlerFunc {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.Write(testBody)
	})
}

func TestCORSMiddleware(t *testing.T) {
	newReq := func(method string, origin string, corsHost string) (req *http.Request) {
		tenantConfig := config.TenantConfiguration{
			UserConfig: config.UserConfiguration{
				CORS: &config.CORSConfiguration{
					Origin: corsHost,
				},
			},
		}
		req, _ = http.NewRequest(method, "", nil)
		if origin != "" {
			req.Header.Set("Origin", origin)
		}
		req = req.WithContext(config.WithTenantConfig(req.Context(), &tenantConfig))
		return
	}

	targetMiddleware := CORSMiddleware{}
	handler := targetMiddleware.Handle(getTestHandler())

	Convey("Test CORSMiddleware", t, func() {
		Convey("should not handle request when CORS config is invalid", func() {
			req := newReq("OPTIONS", "http://test.example.com", "example.*")
			resp := httptest.NewRecorder()
			handler.ServeHTTP(resp, req)

			So(resp.Header().Get("Vary"), ShouldEqual, "Origin")
			So(resp.Header().Get("Access-Control-Allow-Origin"), ShouldBeEmpty)
			So(resp.Header().Get("Access-Control-Allow-Credentials"), ShouldBeEmpty)
		})

		Convey("should handle OPTIONS request", func() {
			req := newReq("OPTIONS", "http://test.example.com", "*.example.com")
			resp := httptest.NewRecorder()
			handler.ServeHTTP(resp, req)

			So(resp.Header().Get("Vary"), ShouldEqual, "Origin")
			So(resp.Header().Get("Access-Control-Allow-Origin"), ShouldEqual, "http://test.example.com")
			So(resp.Header().Get("Access-Control-Allow-Credentials"), ShouldEqual, "true")
			So(resp.Body.Len(), ShouldEqual, 0)
		})

		Convey("should handle localhost", func() {
			req := newReq("OPTIONS", "http://localhost:3000", "localhost:3000")
			resp := httptest.NewRecorder()
			handler.ServeHTTP(resp, req)

			So(resp.Header().Get("Vary"), ShouldEqual, "Origin")
			So(resp.Header().Get("Access-Control-Allow-Origin"), ShouldEqual, "http://localhost:3000")
			So(resp.Header().Get("Access-Control-Allow-Credentials"), ShouldEqual, "true")
			So(resp.Body.Len(), ShouldEqual, 0)
		})

		Convey("should handle POST request", func() {
			req := newReq("POST", "http://test.example.com", "*.example.com")
			resp := httptest.NewRecorder()
			handler.ServeHTTP(resp, req)

			So(resp.Header().Get("Vary"), ShouldEqual, "Origin")
			So(resp.Header().Get("Access-Control-Allow-Origin"), ShouldEqual, "http://test.example.com")
			So(resp.Header().Get("Access-Control-Allow-Credentials"), ShouldEqual, "true")
			So(resp.Body.Bytes(), ShouldResemble, testBody)
		})

		Convey("should handle request with request methods/headers", func() {
			req := newReq("OPTIONS", "http://test.example.com", "*.example.com")
			req.Header.Set("Access-Control-Request-Method", "GET")
			req.Header.Set("Access-Control-Request-Headers", "Cookie")
			resp := httptest.NewRecorder()
			handler.ServeHTTP(resp, req)

			So(resp.Header().Get("Access-Control-Allow-Methods"), ShouldEqual, "GET")
			So(resp.Header().Get("Access-Control-Allow-Headers"), ShouldEqual, "Cookie")
		})

		Convey("should echo request origin as allowed origin", func() {
			req := newReq("OPTIONS", "https://example.com", "*")
			resp := httptest.NewRecorder()
			handler.ServeHTTP(resp, req)

			So(resp.Header().Get("Vary"), ShouldEqual, "Origin")
			So(resp.Header().Get("Access-Control-Allow-Origin"), ShouldEqual, "https://example.com")
			So(resp.Header().Get("Access-Control-Allow-Credentials"), ShouldEqual, "true")
		})

		Convey("should not handle request with not allowed origin", func() {
			req := newReq("OPTIONS", "http://example1.com", "*.example.com")
			resp := httptest.NewRecorder()
			handler.ServeHTTP(resp, req)

			So(resp.Header().Get("Vary"), ShouldEqual, "Origin")
			So(resp.Header().Get("Access-Control-Allow-Origin"), ShouldBeEmpty)
			So(resp.Header().Get("Access-Control-Allow-Credentials"), ShouldBeEmpty)
		})
	})
}
