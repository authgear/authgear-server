package webapp

import (
	"net/http"
	"net/http/httptest"
	"testing"

	gomock "github.com/golang/mock/gomock"
	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/lib/config"
)

func TestDynamicCSPMiddleware(t *testing.T) {
	Convey("DynamicCSPMiddleware", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		cookieManager := NewMockCookieManager(ctrl)
		cookieManager.EXPECT().GetCookie(gomock.Any(), CSPNonceCookieDef).Return(&http.Cookie{}, nil).AnyTimes()

		type TestCase struct {
			OAuthConfig         *config.OAuthConfig
			AllowInlineScript   bool
			AllowFrameAncestors bool
			ExpectedHeaders     map[string][]string
		}

		Convey("CSP Directives", func() {
			testcases := []TestCase{
				{
					OAuthConfig: &config.OAuthConfig{
						Clients: []config.OAuthClientConfig{
							{
								CustomUIURI: "http://customui.com",
							},
						},
					},
					AllowInlineScript:   true,
					AllowFrameAncestors: true,
					ExpectedHeaders: map[string][]string{
						"Content-Security-Policy": {
							"default-src 'self'; script-src 'unsafe-inline' www.googletagmanager.com eu-assets.i.posthog.com https://browser.sentry-cdn.com 'self' http://cdn.authgear.com; frame-src www.googletagmanager.com 'self'; font-src cdnjs.cloudflare.com static2.sharepointonline.com fonts.googleapis.com fonts.gstatic.com 'self' http://cdn.authgear.com; style-src 'unsafe-inline' cdnjs.cloudflare.com www.googletagmanager.com fonts.googleapis.com 'self' http://cdn.authgear.com; img-src http: https: data: 'self' http://cdn.authgear.com; object-src 'none'; base-uri 'none'; connect-src 'self' https://www.google-analytics.com ws://authgear.com wss://authgear.com; block-all-mixed-content; frame-ancestors http://customui.com",
						},
					},
				},
				{
					OAuthConfig: &config.OAuthConfig{
						Clients: []config.OAuthClientConfig{
							{
								CustomUIURI: "http://customui.com",
							},
						},
					},
					AllowInlineScript:   true,
					AllowFrameAncestors: false,
					ExpectedHeaders: map[string][]string{
						"Content-Security-Policy": {
							"default-src 'self'; script-src 'unsafe-inline' www.googletagmanager.com eu-assets.i.posthog.com https://browser.sentry-cdn.com 'self' http://cdn.authgear.com; frame-src www.googletagmanager.com 'self'; font-src cdnjs.cloudflare.com static2.sharepointonline.com fonts.googleapis.com fonts.gstatic.com 'self' http://cdn.authgear.com; style-src 'unsafe-inline' cdnjs.cloudflare.com www.googletagmanager.com fonts.googleapis.com 'self' http://cdn.authgear.com; img-src http: https: data: 'self' http://cdn.authgear.com; object-src 'none'; base-uri 'none'; connect-src 'self' https://www.google-analytics.com ws://authgear.com wss://authgear.com; block-all-mixed-content; frame-ancestors 'none'",
						},
						"X-Frame-Options": {"DENY"},
					},
				},
				{
					OAuthConfig:         &config.OAuthConfig{},
					AllowInlineScript:   true,
					AllowFrameAncestors: true,
					ExpectedHeaders: map[string][]string{
						"Content-Security-Policy": {
							"default-src 'self'; script-src 'unsafe-inline' www.googletagmanager.com eu-assets.i.posthog.com https://browser.sentry-cdn.com 'self' http://cdn.authgear.com; frame-src www.googletagmanager.com 'self'; font-src cdnjs.cloudflare.com static2.sharepointonline.com fonts.googleapis.com fonts.gstatic.com 'self' http://cdn.authgear.com; style-src 'unsafe-inline' cdnjs.cloudflare.com www.googletagmanager.com fonts.googleapis.com 'self' http://cdn.authgear.com; img-src http: https: data: 'self' http://cdn.authgear.com; object-src 'none'; base-uri 'none'; connect-src 'self' https://www.google-analytics.com ws://authgear.com wss://authgear.com; block-all-mixed-content; frame-ancestors 'none'",
						},
						"X-Frame-Options": {"DENY"},
					},
				},
				{
					OAuthConfig:         &config.OAuthConfig{},
					AllowInlineScript:   true,
					AllowFrameAncestors: false,
					ExpectedHeaders: map[string][]string{
						"Content-Security-Policy": {
							"default-src 'self'; script-src 'unsafe-inline' www.googletagmanager.com eu-assets.i.posthog.com https://browser.sentry-cdn.com 'self' http://cdn.authgear.com; frame-src www.googletagmanager.com 'self'; font-src cdnjs.cloudflare.com static2.sharepointonline.com fonts.googleapis.com fonts.gstatic.com 'self' http://cdn.authgear.com; style-src 'unsafe-inline' cdnjs.cloudflare.com www.googletagmanager.com fonts.googleapis.com 'self' http://cdn.authgear.com; img-src http: https: data: 'self' http://cdn.authgear.com; object-src 'none'; base-uri 'none'; connect-src 'self' https://www.google-analytics.com ws://authgear.com wss://authgear.com; block-all-mixed-content; frame-ancestors 'none'",
						},
						"X-Frame-Options": {"DENY"},
					},
				},
				{
					OAuthConfig:         &config.OAuthConfig{},
					AllowInlineScript:   false,
					AllowFrameAncestors: true,
					ExpectedHeaders: map[string][]string{
						"Content-Security-Policy": {
							"default-src 'self'; script-src 'strict-dynamic' 'nonce-' www.googletagmanager.com eu-assets.i.posthog.com https://browser.sentry-cdn.com 'self' http://cdn.authgear.com; frame-src www.googletagmanager.com 'self'; font-src cdnjs.cloudflare.com static2.sharepointonline.com fonts.googleapis.com fonts.gstatic.com 'self' http://cdn.authgear.com; style-src 'unsafe-inline' cdnjs.cloudflare.com www.googletagmanager.com fonts.googleapis.com 'self' http://cdn.authgear.com; img-src http: https: data: 'self' http://cdn.authgear.com; object-src 'none'; base-uri 'none'; connect-src 'self' https://www.google-analytics.com ws://authgear.com wss://authgear.com; block-all-mixed-content; frame-ancestors 'none'",
						},
						"X-Frame-Options": {"DENY"},
					},
				},
				{
					OAuthConfig:         &config.OAuthConfig{},
					AllowInlineScript:   false,
					AllowFrameAncestors: false,
					ExpectedHeaders: map[string][]string{
						"Content-Security-Policy": {
							"default-src 'self'; script-src 'strict-dynamic' 'nonce-' www.googletagmanager.com eu-assets.i.posthog.com https://browser.sentry-cdn.com 'self' http://cdn.authgear.com; frame-src www.googletagmanager.com 'self'; font-src cdnjs.cloudflare.com static2.sharepointonline.com fonts.googleapis.com fonts.gstatic.com 'self' http://cdn.authgear.com; style-src 'unsafe-inline' cdnjs.cloudflare.com www.googletagmanager.com fonts.googleapis.com 'self' http://cdn.authgear.com; img-src http: https: data: 'self' http://cdn.authgear.com; object-src 'none'; base-uri 'none'; connect-src 'self' https://www.google-analytics.com ws://authgear.com wss://authgear.com; block-all-mixed-content; frame-ancestors 'none'",
						},
						"X-Frame-Options": {"DENY"},
					},
				},
			}

			test := func(testcase TestCase, resultHeaders map[string][]string) {
				middleware := DynamicCSPMiddleware{
					Cookies:             cookieManager,
					HTTPOrigin:          "http://authgear.com",
					WebAppCDNHost:       "http://cdn.authgear.com",
					OAuthConfig:         testcase.OAuthConfig,
					AllowInlineScript:   AllowInlineScript(testcase.AllowInlineScript),
					AllowFrameAncestors: AllowFrameAncestors(testcase.AllowFrameAncestors),
				}

				makeHandler := func() http.Handler {
					dummy := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
					h := middleware.Handle(dummy)
					return h
				}

				w := httptest.NewRecorder()
				r, _ := http.NewRequest("GET", "/", nil)
				makeHandler().ServeHTTP(w, r)

				for key, values := range resultHeaders {
					So(w.Header()[key], ShouldResemble, values)
				}

				for key := range w.Header() {
					if _, ok := resultHeaders[key]; !ok {
						t.Errorf("unexpected header: %s", key)
					}
				}
			}

			for _, testcase := range testcases {
				test(testcase, testcase.ExpectedHeaders)
			}
		})
	})
}
