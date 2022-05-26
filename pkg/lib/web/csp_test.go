package web

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestCSPJoin(t *testing.T) {
	Convey("CSPJoin", t, func() {
		So(CSPJoin([]string{"a", "b"}), ShouldResemble, "a; b")
	})
}

func TestCSPDirectives(t *testing.T) {
	Convey("CSPDirectives", t, func() {
		test := func(publicOrigin string, nonce string, expected []string) {
			actual, err := CSPDirectives(publicOrigin, nonce)
			So(err, ShouldBeNil)
			So(actual, ShouldResemble, expected)
		}

		test("http://localhost:3000", "N0NC5", []string{
			"default-src 'self'",
			"script-src 'self' 'nonce-N0NC5' www.googletagmanager.com",
			"frame-src 'self' www.googletagmanager.com",
			"font-src 'self' cdnjs.cloudflare.com static2.sharepointonline.com fonts.googleapis.com fonts.gstatic.com",
			"style-src 'self' 'unsafe-inline' cdnjs.cloudflare.com www.googletagmanager.com fonts.googleapis.com",
			"img-src 'self' http: https: data:",
			"object-src 'none'",
			"base-uri 'none'",
			"connect-src 'self' https://www.google-analytics.com ws://localhost:3000 wss://localhost:3000",
			"block-all-mixed-content",
			"frame-ancestors 'none'",
		})
	})
}
