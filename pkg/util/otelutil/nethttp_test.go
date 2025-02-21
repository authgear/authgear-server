package otelutil

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/felixge/httpsnoop"
	. "github.com/smartystreets/goconvey/convey"
)

func TestHTTPRequestMethod(t *testing.T) {
	Convey("HTTPRequestMethod", t, func() {
		test := func(method string, expected string) {
			r := &http.Request{Method: method}
			actual := HTTPRequestMethod(r)
			So(actual.Value.AsString(), ShouldEqual, expected)
			So(string(actual.Key), ShouldEqual, "http.request.method")
		}

		test("GET", "GET")
		test("HEAD", "HEAD")
		test("POST", "POST")
		test("PUT", "PUT")
		test("DELETE", "DELETE")
		test("CONNECT", "CONNECT")
		test("OPTIONS", "OPTIONS")
		test("TRACE", "TRACE")
		test("PATCH", "PATCH")

		test("", "_OTHER")
		test("nonsense", "_OTHER")
	})
}

func TestHTTPURLScheme(t *testing.T) {
	Convey("HTTPURLScheme", t, func() {
		test := func(scheme string, expected string) {
			actual := HTTPURLScheme(scheme)
			So(actual.Value.AsString(), ShouldEqual, expected)
			So(string(actual.Key), ShouldEqual, "url.scheme")
		}

		test("http", "http")
		test("https", "https")

		test("telnet", "http")
	})
}

func TestHTTPResponseStatusCode(t *testing.T) {
	Convey("HTTPResponseStatusCode", t, func() {
		test := func(h http.Handler, expected int) {
			ctx := context.Background()
			r := httptest.NewRequestWithContext(ctx, "GET", "/", nil)
			w := httptest.NewRecorder()

			metrics := httpsnoop.CaptureMetrics(h, w, r)
			actual := HTTPResponseStatusCode(metrics)

			So(actual.Value.AsInt64(), ShouldEqual, expected)
			So(string(actual.Key), ShouldEqual, "http.response.status_code")
		}

		test(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}), 200)
		test(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(201)
		}), 201)
		test(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, "", http.StatusFound)
		}), 302)
	})
}
