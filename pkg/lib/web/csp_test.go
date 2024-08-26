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
			So(actual, ShouldResemble, expected)
		}

		test(CSPDirectivesOptions{
			PublicOrigin:      "http://localhost:3000",
			Nonce:             "N0NC5",
			CDNHost:           "",
			AllowInlineScript: false,
		}, []string{
			"default-src 'self'",
			"script-src 'strict-dynamic' 'nonce-N0NC5' eu-assets.i.posthog.com https://browser.sentry-cdn.com 'self'",
			"frame-src www.googletagmanager.com challenges.cloudflare.com www.google.com 'self'",
			"font-src cdnjs.cloudflare.com static2.sharepointonline.com fonts.googleapis.com fonts.gstatic.com 'self'",
			"style-src 'unsafe-inline' cdnjs.cloudflare.com www.googletagmanager.com fonts.googleapis.com 'self'",
			"img-src http: https: data: 'self'",
			"object-src 'none'",
			"base-uri 'none'",
			"connect-src 'self' https://www.google-analytics.com ws://localhost:3000 wss://localhost:3000",
			"block-all-mixed-content",
			"frame-ancestors 'none'",
		})

		test(CSPDirectivesOptions{
			PublicOrigin:      "http://localhost:3000",
			Nonce:             "N0NC5",
			CDNHost:           "cdn.localhost:3000",
			AllowInlineScript: false,
		}, []string{
			"default-src 'self'",
			"script-src 'strict-dynamic' 'nonce-N0NC5' eu-assets.i.posthog.com https://browser.sentry-cdn.com 'self' cdn.localhost:3000",
			"frame-src www.googletagmanager.com challenges.cloudflare.com www.google.com 'self'",
			"font-src cdnjs.cloudflare.com static2.sharepointonline.com fonts.googleapis.com fonts.gstatic.com 'self' cdn.localhost:3000",
			"style-src 'unsafe-inline' cdnjs.cloudflare.com www.googletagmanager.com fonts.googleapis.com 'self' cdn.localhost:3000",
			"img-src http: https: data: 'self' cdn.localhost:3000",
			"object-src 'none'",
			"base-uri 'none'",
			"connect-src 'self' https://www.google-analytics.com ws://localhost:3000 wss://localhost:3000",
			"block-all-mixed-content",
			"frame-ancestors 'none'",
		})

		test(CSPDirectivesOptions{
			PublicOrigin:      "http://localhost:3000",
			Nonce:             "N0NC5",
			CDNHost:           "cdn.localhost:3000",
			AllowInlineScript: true,
		}, []string{
			"default-src 'self'",
			"script-src 'unsafe-inline' eu-assets.i.posthog.com https://browser.sentry-cdn.com 'self' cdn.localhost:3000",
			"frame-src www.googletagmanager.com challenges.cloudflare.com www.google.com 'self'",
			"font-src cdnjs.cloudflare.com static2.sharepointonline.com fonts.googleapis.com fonts.gstatic.com 'self' cdn.localhost:3000",
			"style-src 'unsafe-inline' cdnjs.cloudflare.com www.googletagmanager.com fonts.googleapis.com 'self' cdn.localhost:3000",
			"img-src http: https: data: 'self' cdn.localhost:3000",
			"object-src 'none'",
			"base-uri 'none'",
			"connect-src 'self' https://www.google-analytics.com ws://localhost:3000 wss://localhost:3000",
			"block-all-mixed-content",
			"frame-ancestors 'none'",
		})

		test(CSPDirectivesOptions{
			PublicOrigin:      "http://localhost:3000",
			Nonce:             "N0NC5",
			CDNHost:           "cdn.localhost:3000",
			AllowInlineScript: false,
			FrameAncestors:    []string{"http://remote.localhost"},
		}, []string{
			"default-src 'self'",
			"script-src 'strict-dynamic' 'nonce-N0NC5' eu-assets.i.posthog.com https://browser.sentry-cdn.com 'self' cdn.localhost:3000",
			"frame-src www.googletagmanager.com challenges.cloudflare.com www.google.com 'self'",
			"font-src cdnjs.cloudflare.com static2.sharepointonline.com fonts.googleapis.com fonts.gstatic.com 'self' cdn.localhost:3000",
			"style-src 'unsafe-inline' cdnjs.cloudflare.com www.googletagmanager.com fonts.googleapis.com 'self' cdn.localhost:3000",
			"img-src http: https: data: 'self' cdn.localhost:3000",
			"object-src 'none'",
			"base-uri 'none'",
			"connect-src 'self' https://www.google-analytics.com ws://localhost:3000 wss://localhost:3000",
			"block-all-mixed-content",
			"frame-ancestors http://remote.localhost",
		})

		test(CSPDirectivesOptions{
			PublicOrigin:      "http://localhost:3000",
			Nonce:             "N0NC5",
			CDNHost:           "",
			AllowInlineScript: false,
			AuthUISentryDSN:   "https://examplePublicKey@o0.ingest.sentry.io/0",
		}, []string{
			"default-src 'self'",
			"script-src 'strict-dynamic' 'nonce-N0NC5' eu-assets.i.posthog.com https://browser.sentry-cdn.com 'self'",
			"frame-src www.googletagmanager.com challenges.cloudflare.com www.google.com 'self'",
			"font-src cdnjs.cloudflare.com static2.sharepointonline.com fonts.googleapis.com fonts.gstatic.com 'self'",
			"style-src 'unsafe-inline' cdnjs.cloudflare.com www.googletagmanager.com fonts.googleapis.com 'self'",
			"img-src http: https: data: 'self'",
			"object-src 'none'",
			"base-uri 'none'",
			"connect-src 'self' https://www.google-analytics.com ws://localhost:3000 wss://localhost:3000 o0.ingest.sentry.io",
			"block-all-mixed-content",
			"frame-ancestors 'none'",
		})
	})
}
