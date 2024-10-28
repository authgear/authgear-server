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
			CDNHost:      "",
		}, []string{
			"default-src 'self'",
			"script-src 'self' www.googletagmanager.com eu-assets.i.posthog.com challenges.cloudflare.com www.google.com https://browser.sentry-cdn.com 'nonce-N0NC5' 'strict-dynamic'",
			"frame-src 'self' www.googletagmanager.com challenges.cloudflare.com www.google.com",
			"font-src 'self' cdnjs.cloudflare.com static2.sharepointonline.com fonts.googleapis.com fonts.gstatic.com",
			"style-src 'self' cdnjs.cloudflare.com www.googletagmanager.com fonts.googleapis.com 'sha256-WAyOw4V+FqDc35lQPyRADLBWbuNK8ahvYEaQIYF1+Ps=' 'sha256-fOghyYcDMsLl/lf7piKeVgEljdV7IgqwGymlDo5oDhU=' 'sha256-0EZqoz+oBhx7gF4nvY2bSqoGyy4zLjNF+SDQXGp/ZrY=' 'sha256-ZLjZaRfcYelvFE+8S7ynGAe0XPN7SLX6dirEzdvD5Mk=' 'unsafe-hashes' 'nonce-N0NC5'",
			"img-src 'self' http: https: data:",
			"object-src 'none'",
			"base-uri 'none'",
			"connect-src 'self' https://www.google-analytics.com ws://localhost:3000 wss://localhost:3000",
			"block-all-mixed-content",
			"frame-ancestors 'none'",
		})

		test(CSPDirectivesOptions{
			PublicOrigin: "http://localhost:3000",
			Nonce:        "N0NC5",
			CDNHost:      "cdn.localhost:3000",
		}, []string{
			"default-src 'self'",
			"script-src 'self' www.googletagmanager.com eu-assets.i.posthog.com challenges.cloudflare.com www.google.com https://browser.sentry-cdn.com cdn.localhost:3000 'nonce-N0NC5' 'strict-dynamic'",
			"frame-src 'self' www.googletagmanager.com challenges.cloudflare.com www.google.com",
			"font-src 'self' cdnjs.cloudflare.com static2.sharepointonline.com fonts.googleapis.com fonts.gstatic.com cdn.localhost:3000",
			"style-src 'self' cdnjs.cloudflare.com www.googletagmanager.com fonts.googleapis.com cdn.localhost:3000 'sha256-WAyOw4V+FqDc35lQPyRADLBWbuNK8ahvYEaQIYF1+Ps=' 'sha256-fOghyYcDMsLl/lf7piKeVgEljdV7IgqwGymlDo5oDhU=' 'sha256-0EZqoz+oBhx7gF4nvY2bSqoGyy4zLjNF+SDQXGp/ZrY=' 'sha256-ZLjZaRfcYelvFE+8S7ynGAe0XPN7SLX6dirEzdvD5Mk=' 'unsafe-hashes' 'nonce-N0NC5'",
			"img-src 'self' http: https: data: cdn.localhost:3000",
			"object-src 'none'",
			"base-uri 'none'",
			"connect-src 'self' https://www.google-analytics.com ws://localhost:3000 wss://localhost:3000",
			"block-all-mixed-content",
			"frame-ancestors 'none'",
		})

		test(CSPDirectivesOptions{
			PublicOrigin: "http://localhost:3000",
			Nonce:        "N0NC5",
			CDNHost:      "cdn.localhost:3000",
		}, []string{
			"default-src 'self'",
			"script-src 'self' www.googletagmanager.com eu-assets.i.posthog.com challenges.cloudflare.com www.google.com https://browser.sentry-cdn.com cdn.localhost:3000 'nonce-N0NC5' 'strict-dynamic'",
			"frame-src 'self' www.googletagmanager.com challenges.cloudflare.com www.google.com",
			"font-src 'self' cdnjs.cloudflare.com static2.sharepointonline.com fonts.googleapis.com fonts.gstatic.com cdn.localhost:3000",
			"style-src 'self' cdnjs.cloudflare.com www.googletagmanager.com fonts.googleapis.com cdn.localhost:3000 'sha256-WAyOw4V+FqDc35lQPyRADLBWbuNK8ahvYEaQIYF1+Ps=' 'sha256-fOghyYcDMsLl/lf7piKeVgEljdV7IgqwGymlDo5oDhU=' 'sha256-0EZqoz+oBhx7gF4nvY2bSqoGyy4zLjNF+SDQXGp/ZrY=' 'sha256-ZLjZaRfcYelvFE+8S7ynGAe0XPN7SLX6dirEzdvD5Mk=' 'unsafe-hashes' 'nonce-N0NC5'",
			"img-src 'self' http: https: data: cdn.localhost:3000",
			"object-src 'none'",
			"base-uri 'none'",
			"connect-src 'self' https://www.google-analytics.com ws://localhost:3000 wss://localhost:3000",
			"block-all-mixed-content",
			"frame-ancestors 'none'",
		})

		test(CSPDirectivesOptions{
			PublicOrigin:   "http://localhost:3000",
			Nonce:          "N0NC5",
			CDNHost:        "cdn.localhost:3000",
			FrameAncestors: []string{"http://remote.localhost"},
		}, []string{
			"default-src 'self'",
			"script-src 'self' www.googletagmanager.com eu-assets.i.posthog.com challenges.cloudflare.com www.google.com https://browser.sentry-cdn.com cdn.localhost:3000 'nonce-N0NC5' 'strict-dynamic'",
			"frame-src 'self' www.googletagmanager.com challenges.cloudflare.com www.google.com",
			"font-src 'self' cdnjs.cloudflare.com static2.sharepointonline.com fonts.googleapis.com fonts.gstatic.com cdn.localhost:3000",
			"style-src 'self' cdnjs.cloudflare.com www.googletagmanager.com fonts.googleapis.com cdn.localhost:3000 'sha256-WAyOw4V+FqDc35lQPyRADLBWbuNK8ahvYEaQIYF1+Ps=' 'sha256-fOghyYcDMsLl/lf7piKeVgEljdV7IgqwGymlDo5oDhU=' 'sha256-0EZqoz+oBhx7gF4nvY2bSqoGyy4zLjNF+SDQXGp/ZrY=' 'sha256-ZLjZaRfcYelvFE+8S7ynGAe0XPN7SLX6dirEzdvD5Mk=' 'unsafe-hashes' 'nonce-N0NC5'",
			"img-src 'self' http: https: data: cdn.localhost:3000",
			"object-src 'none'",
			"base-uri 'none'",
			"connect-src 'self' https://www.google-analytics.com ws://localhost:3000 wss://localhost:3000",
			"block-all-mixed-content",
			"frame-ancestors http://remote.localhost",
		})

		test(CSPDirectivesOptions{
			PublicOrigin:    "http://localhost:3000",
			Nonce:           "N0NC5",
			CDNHost:         "",
			AuthUISentryDSN: "https://examplePublicKey@o0.ingest.sentry.io/0",
		}, []string{
			"default-src 'self'",
			"script-src 'self' www.googletagmanager.com eu-assets.i.posthog.com challenges.cloudflare.com www.google.com https://browser.sentry-cdn.com 'nonce-N0NC5' 'strict-dynamic'",
			"frame-src 'self' www.googletagmanager.com challenges.cloudflare.com www.google.com",
			"font-src 'self' cdnjs.cloudflare.com static2.sharepointonline.com fonts.googleapis.com fonts.gstatic.com",
			"style-src 'self' cdnjs.cloudflare.com www.googletagmanager.com fonts.googleapis.com 'sha256-WAyOw4V+FqDc35lQPyRADLBWbuNK8ahvYEaQIYF1+Ps=' 'sha256-fOghyYcDMsLl/lf7piKeVgEljdV7IgqwGymlDo5oDhU=' 'sha256-0EZqoz+oBhx7gF4nvY2bSqoGyy4zLjNF+SDQXGp/ZrY=' 'sha256-ZLjZaRfcYelvFE+8S7ynGAe0XPN7SLX6dirEzdvD5Mk=' 'unsafe-hashes' 'nonce-N0NC5'",
			"img-src 'self' http: https: data:",
			"object-src 'none'",
			"base-uri 'none'",
			"connect-src 'self' https://www.google-analytics.com ws://localhost:3000 wss://localhost:3000 o0.ingest.sentry.io",
			"block-all-mixed-content",
			"frame-ancestors 'none'",
		})
	})
}
