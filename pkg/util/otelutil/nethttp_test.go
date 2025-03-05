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

func TestHTTPServerAddress(t *testing.T) {
	Convey("HTTPServerAddress", t, func() {
		test := func(hostAndPort string, expected string) {
			actual, ok := HTTPServerAddress(hostAndPort)
			if expected == "" {
				So(ok, ShouldBeFalse)
				So(actual, ShouldBeNil)
			} else {
				So(actual.Value.AsString(), ShouldEqual, expected)
				So(string(actual.Key), ShouldEqual, "server.address")
			}
		}

		// Lookback
		test("localhost", "localhost")
		test("127.0.0.1", "127.0.0.1")
		test("[::1]", "::1")
		test("localhost:80", "localhost")
		test("127.0.0.1:80", "127.0.0.1")
		test("[::1]:80", "::1")

		// Real world
		test("example.com", "example.com")
		test("example.com:80", "example.com")
		test("example.com:443", "example.com")

		// Missing ]
		test("[::1", "")
		// IPv6 not enclosed in []
		test("::1:80", "")
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
