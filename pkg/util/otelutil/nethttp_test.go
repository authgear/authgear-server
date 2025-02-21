package otelutil

import (
	"net/http"
	"testing"

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
		test := func(statusCode int) {
			actual := HTTPResponseStatusCode(statusCode)

			So(actual.Value.AsInt64(), ShouldEqual, statusCode)
			So(string(actual.Key), ShouldEqual, "http.response.status_code")
		}

		test(200)
		test(201)
		test(302)
	})
}
