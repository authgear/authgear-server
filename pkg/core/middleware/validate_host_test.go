package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestValidateHostMiddleware(t *testing.T) {
	makeHandler := func(validHosts string) http.Handler {
		targetMiddleware := ValidateHostMiddleware{
			ValidHosts: validHosts,
		}
		return targetMiddleware.Handle(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			rw.WriteHeader(http.StatusOK)
		}))
	}
	var req *http.Request
	var resp *httptest.ResponseRecorder

	Convey("Test ValidateHostMiddleware", t, func() {
		Convey("should allow correct hosts", func() {
			handler := makeHandler("localhost, 127.0.0.1")

			req, _ = http.NewRequest("GET", "https://localhost/test", nil)
			resp = httptest.NewRecorder()
			handler.ServeHTTP(resp, req)
			So(resp.Code, ShouldEqual, http.StatusOK)

			req, _ = http.NewRequest("GET", "http://127.0.0.1/test", nil)
			resp = httptest.NewRecorder()
			handler.ServeHTTP(resp, req)
			So(resp.Code, ShouldEqual, http.StatusOK)

			req, _ = http.NewRequest("GET", "https://example.com/test", nil)
			resp = httptest.NewRecorder()
			handler.ServeHTTP(resp, req)
			So(resp.Code, ShouldEqual, http.StatusBadRequest)
		})

		Convey("should check ports", func() {
			handler := makeHandler("127.0.0.1:3000")

			req, _ = http.NewRequest("GET", "https://127.0.0.1/test", nil)
			resp = httptest.NewRecorder()
			handler.ServeHTTP(resp, req)
			So(resp.Code, ShouldEqual, http.StatusBadRequest)

			req, _ = http.NewRequest("GET", "http://127.0.0.1:3000/", nil)
			resp = httptest.NewRecorder()
			handler.ServeHTTP(resp, req)
			So(resp.Code, ShouldEqual, http.StatusOK)
		})

		Convey("should check X-Forwarded-Host header", func() {
			handler := makeHandler("example.com")

			req, _ = http.NewRequest("GET", "https://gateway.com/test", nil)
			req.Header.Set("X-Forwarded-Host", "example.com")
			resp = httptest.NewRecorder()
			handler.ServeHTTP(resp, req)
			So(resp.Code, ShouldEqual, http.StatusOK)

			req, _ = http.NewRequest("GET", "https://example.com/test", nil)
			req.Header.Set("X-Forwarded-Host", "internal.com")
			resp = httptest.NewRecorder()
			handler.ServeHTTP(resp, req)
			So(resp.Code, ShouldEqual, http.StatusBadRequest)
		})

		Convey("should skip validating hosts", func() {
			handler := makeHandler(" ")

			req, _ = http.NewRequest("GET", "https://localhost/test", nil)
			resp = httptest.NewRecorder()
			handler.ServeHTTP(resp, req)
			So(resp.Code, ShouldEqual, http.StatusOK)
		})
	})
}
