package web

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestCSPDirectives(t *testing.T) {
	Convey("CSPDirectives", t, func() {
		test := func(opts CSPDirectivesOptions, expected []string) {
			actual, err := CSPDirectives(opts)
			So(err, ShouldBeNil)
			var strs []string
			for _, directive := range actual {
				strs = append(strs, directive.String())
			}
			So(strs, ShouldResemble, expected)
		}

		test(CSPDirectivesOptions{
			PublicOrigin: "http://localhost:3000",
			Nonce:        "N0NC5",
		}, []string{
			"default-src 'self'",
			"script-src 'self' https: 'nonce-N0NC5' 'strict-dynamic'",
			"frame-src 'self' https:",
			"font-src 'self' https:",
			"style-src 'self' 'unsafe-inline' 'unsafe-eval' https:",
			"img-src 'self' http: https: data:",
			"object-src 'none'",
			"base-uri 'none'",
			"connect-src 'self' https: ws://localhost:3000 wss://localhost:3000",
			"block-all-mixed-content",
			"frame-ancestors 'none'",
		})

		test(CSPDirectivesOptions{
			PublicOrigin: "http://localhost:3000",
			Nonce:        "N0NC5",
		}, []string{
			"default-src 'self'",
			"script-src 'self' https: 'nonce-N0NC5' 'strict-dynamic'",
			"frame-src 'self' https:",
			"font-src 'self' https:",
			"style-src 'self' 'unsafe-inline' 'unsafe-eval' https:",
			"img-src 'self' http: https: data:",
			"object-src 'none'",
			"base-uri 'none'",
			"connect-src 'self' https: ws://localhost:3000 wss://localhost:3000",
			"block-all-mixed-content",
			"frame-ancestors 'none'",
		})

		test(CSPDirectivesOptions{
			PublicOrigin: "http://localhost:3000",
			Nonce:        "N0NC5",
		}, []string{
			"default-src 'self'",
			"script-src 'self' https: 'nonce-N0NC5' 'strict-dynamic'",
			"frame-src 'self' https:",
			"font-src 'self' https:",
			"style-src 'self' 'unsafe-inline' 'unsafe-eval' https:",
			"img-src 'self' http: https: data:",
			"object-src 'none'",
			"base-uri 'none'",
			"connect-src 'self' https: ws://localhost:3000 wss://localhost:3000",
			"block-all-mixed-content",
			"frame-ancestors 'none'",
		})

		test(CSPDirectivesOptions{
			PublicOrigin:   "http://localhost:3000",
			Nonce:          "N0NC5",
			FrameAncestors: []string{"http://remote.localhost"},
		}, []string{
			"default-src 'self'",
			"script-src 'self' https: 'nonce-N0NC5' 'strict-dynamic'",
			"frame-src 'self' https:",
			"font-src 'self' https:",
			"style-src 'self' 'unsafe-inline' 'unsafe-eval' https:",
			"img-src 'self' http: https: data:",
			"object-src 'none'",
			"base-uri 'none'",
			"connect-src 'self' https: ws://localhost:3000 wss://localhost:3000",
			"block-all-mixed-content",
			"frame-ancestors http://remote.localhost",
		})

		test(CSPDirectivesOptions{
			PublicOrigin: "http://localhost:3000",
			Nonce:        "N0NC5",
		}, []string{
			"default-src 'self'",
			"script-src 'self' https: 'nonce-N0NC5' 'strict-dynamic'",
			"frame-src 'self' https:",
			"font-src 'self' https:",
			"style-src 'self' 'unsafe-inline' 'unsafe-eval' https:",
			"img-src 'self' http: https: data:",
			"object-src 'none'",
			"base-uri 'none'",
			"connect-src 'self' https: ws://localhost:3000 wss://localhost:3000",
			"block-all-mixed-content",
			"frame-ancestors 'none'",
		})
	})
}
