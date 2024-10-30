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
			Nonce: "N0NC5",
		}, []string{
			"script-src 'self' https: 'nonce-N0NC5' 'strict-dynamic'",
			"object-src 'none'",
			"base-uri 'none'",
			"frame-ancestors 'none'",
		})

		test(CSPDirectivesOptions{
			Nonce: "N0NC5",
		}, []string{
			"script-src 'self' https: 'nonce-N0NC5' 'strict-dynamic'",
			"object-src 'none'",
			"base-uri 'none'",
			"frame-ancestors 'none'",
		})

		test(CSPDirectivesOptions{
			Nonce: "N0NC5",
		}, []string{
			"script-src 'self' https: 'nonce-N0NC5' 'strict-dynamic'",
			"object-src 'none'",
			"base-uri 'none'",
			"frame-ancestors 'none'",
		})

		test(CSPDirectivesOptions{
			Nonce:          "N0NC5",
			FrameAncestors: []string{"http://remote.localhost"},
		}, []string{
			"script-src 'self' https: 'nonce-N0NC5' 'strict-dynamic'",
			"object-src 'none'",
			"base-uri 'none'",
			"frame-ancestors http://remote.localhost",
		})

		test(CSPDirectivesOptions{
			Nonce: "N0NC5",
		}, []string{
			"script-src 'self' https: 'nonce-N0NC5' 'strict-dynamic'",
			"object-src 'none'",
			"base-uri 'none'",
			"frame-ancestors 'none'",
		})
	})
}
