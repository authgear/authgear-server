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
			AllowedFrameAncestorsFromEnv    config.AllowedFrameAncestors
			OAuthConfig                     *config.OAuthConfig
			AllowInlineScript               bool
			AllowFrameAncestorsFromEnv      bool
			AllowFrameAncestorsFromCustomUI bool
			ExpectedHeaders                 map[string][]string
		}

		Convey("CSP Directives", func() {
			testcases := []TestCase{
				{
					AllowedFrameAncestorsFromEnv: config.AllowedFrameAncestors{"http://authgearportal.com"},
					OAuthConfig: &config.OAuthConfig{
						Clients: []config.OAuthClientConfig{
							{
								CustomUIURI: "http://customui.com",
							},
						},
					},
					AllowInlineScript:               true,
					AllowFrameAncestorsFromEnv:      true,
					AllowFrameAncestorsFromCustomUI: true,
					ExpectedHeaders: map[string][]string{
						"Content-Security-Policy": {
							"default-src 'self'; script-src 'unsafe-inline' www.googletagmanager.com eu-assets.i.posthog.com challenges.cloudflare.com www.google.com https://browser.sentry-cdn.com 'self' http://cdn.authgear.com; frame-src www.googletagmanager.com challenges.cloudflare.com www.google.com 'self'; font-src cdnjs.cloudflare.com static2.sharepointonline.com fonts.googleapis.com fonts.gstatic.com 'self' http://cdn.authgear.com; style-src 'unsafe-inline' cdnjs.cloudflare.com www.googletagmanager.com fonts.googleapis.com 'self' http://cdn.authgear.com; img-src http: https: data: 'self' http://cdn.authgear.com; object-src 'none'; base-uri 'none'; connect-src 'self' https://www.google-analytics.com ws://authgear.com wss://authgear.com; block-all-mixed-content; frame-ancestors http://authgearportal.com http://customui.com",
						},
					},
				},
				{
					AllowedFrameAncestorsFromEnv: config.AllowedFrameAncestors{"http://authgearportal.com"},
					OAuthConfig: &config.OAuthConfig{
						Clients: []config.OAuthClientConfig{
							{
								CustomUIURI: "http://customui.com",
							},
						},
					},
					AllowInlineScript:               true,
					AllowFrameAncestorsFromEnv:      true,
					AllowFrameAncestorsFromCustomUI: false,
					ExpectedHeaders: map[string][]string{
						"Content-Security-Policy": {
							"default-src 'self'; script-src 'unsafe-inline' www.googletagmanager.com eu-assets.i.posthog.com challenges.cloudflare.com www.google.com https://browser.sentry-cdn.com 'self' http://cdn.authgear.com; frame-src www.googletagmanager.com challenges.cloudflare.com www.google.com 'self'; font-src cdnjs.cloudflare.com static2.sharepointonline.com fonts.googleapis.com fonts.gstatic.com 'self' http://cdn.authgear.com; style-src 'unsafe-inline' cdnjs.cloudflare.com www.googletagmanager.com fonts.googleapis.com 'self' http://cdn.authgear.com; img-src http: https: data: 'self' http://cdn.authgear.com; object-src 'none'; base-uri 'none'; connect-src 'self' https://www.google-analytics.com ws://authgear.com wss://authgear.com; block-all-mixed-content; frame-ancestors http://authgearportal.com",
						},
					},
				},
				{
					AllowedFrameAncestorsFromEnv: config.AllowedFrameAncestors{"http://authgearportal.com"},
					OAuthConfig: &config.OAuthConfig{
						Clients: []config.OAuthClientConfig{
							{
								CustomUIURI: "http://customui.com",
							},
						},
					},
					AllowInlineScript:               true,
					AllowFrameAncestorsFromEnv:      false,
					AllowFrameAncestorsFromCustomUI: true,
					ExpectedHeaders: map[string][]string{
						"Content-Security-Policy": {
							"default-src 'self'; script-src 'unsafe-inline' www.googletagmanager.com eu-assets.i.posthog.com challenges.cloudflare.com www.google.com https://browser.sentry-cdn.com 'self' http://cdn.authgear.com; frame-src www.googletagmanager.com challenges.cloudflare.com www.google.com 'self'; font-src cdnjs.cloudflare.com static2.sharepointonline.com fonts.googleapis.com fonts.gstatic.com 'self' http://cdn.authgear.com; style-src 'unsafe-inline' cdnjs.cloudflare.com www.googletagmanager.com fonts.googleapis.com 'self' http://cdn.authgear.com; img-src http: https: data: 'self' http://cdn.authgear.com; object-src 'none'; base-uri 'none'; connect-src 'self' https://www.google-analytics.com ws://authgear.com wss://authgear.com; block-all-mixed-content; frame-ancestors http://customui.com",
						},
					},
				},
				{
					AllowedFrameAncestorsFromEnv: config.AllowedFrameAncestors{"http://authgearportal.com"},
					OAuthConfig: &config.OAuthConfig{
						Clients: []config.OAuthClientConfig{
							{
								CustomUIURI: "http://customui.com",
							},
						},
					},
					AllowInlineScript:               true,
					AllowFrameAncestorsFromEnv:      false,
					AllowFrameAncestorsFromCustomUI: false,
					ExpectedHeaders: map[string][]string{
						"Content-Security-Policy": {
							"default-src 'self'; script-src 'unsafe-inline' www.googletagmanager.com eu-assets.i.posthog.com challenges.cloudflare.com www.google.com https://browser.sentry-cdn.com 'self' http://cdn.authgear.com; frame-src www.googletagmanager.com challenges.cloudflare.com www.google.com 'self'; font-src cdnjs.cloudflare.com static2.sharepointonline.com fonts.googleapis.com fonts.gstatic.com 'self' http://cdn.authgear.com; style-src 'unsafe-inline' cdnjs.cloudflare.com www.googletagmanager.com fonts.googleapis.com 'self' http://cdn.authgear.com; img-src http: https: data: 'self' http://cdn.authgear.com; object-src 'none'; base-uri 'none'; connect-src 'self' https://www.google-analytics.com ws://authgear.com wss://authgear.com; block-all-mixed-content; frame-ancestors 'none'",
						},
						"X-Frame-Options": {"DENY"},
					},
				},
				{
					AllowedFrameAncestorsFromEnv: config.AllowedFrameAncestors{"http://authgearportal.com"},
					OAuthConfig: &config.OAuthConfig{
						Clients: []config.OAuthClientConfig{
							{
								CustomUIURI: "http://customui.com",
							},
						},
					},
					AllowInlineScript:               false,
					AllowFrameAncestorsFromEnv:      true,
					AllowFrameAncestorsFromCustomUI: true,
					ExpectedHeaders: map[string][]string{
						"Content-Security-Policy": {
							"default-src 'self'; script-src 'strict-dynamic' 'nonce-' www.googletagmanager.com eu-assets.i.posthog.com challenges.cloudflare.com www.google.com https://browser.sentry-cdn.com 'self' http://cdn.authgear.com; frame-src www.googletagmanager.com challenges.cloudflare.com www.google.com 'self'; font-src cdnjs.cloudflare.com static2.sharepointonline.com fonts.googleapis.com fonts.gstatic.com 'self' http://cdn.authgear.com; style-src 'unsafe-inline' cdnjs.cloudflare.com www.googletagmanager.com fonts.googleapis.com 'self' http://cdn.authgear.com; img-src http: https: data: 'self' http://cdn.authgear.com; object-src 'none'; base-uri 'none'; connect-src 'self' https://www.google-analytics.com ws://authgear.com wss://authgear.com; block-all-mixed-content; frame-ancestors http://authgearportal.com http://customui.com",
						},
					},
				},
				{
					AllowedFrameAncestorsFromEnv: config.AllowedFrameAncestors{"http://authgearportal.com"},
					OAuthConfig: &config.OAuthConfig{
						Clients: []config.OAuthClientConfig{
							{
								CustomUIURI: "http://customui.com",
							},
						},
					},
					AllowInlineScript:               false,
					AllowFrameAncestorsFromEnv:      true,
					AllowFrameAncestorsFromCustomUI: false,
					ExpectedHeaders: map[string][]string{
						"Content-Security-Policy": {
							"default-src 'self'; script-src 'strict-dynamic' 'nonce-' www.googletagmanager.com eu-assets.i.posthog.com challenges.cloudflare.com www.google.com https://browser.sentry-cdn.com 'self' http://cdn.authgear.com; frame-src www.googletagmanager.com challenges.cloudflare.com www.google.com 'self'; font-src cdnjs.cloudflare.com static2.sharepointonline.com fonts.googleapis.com fonts.gstatic.com 'self' http://cdn.authgear.com; style-src 'unsafe-inline' cdnjs.cloudflare.com www.googletagmanager.com fonts.googleapis.com 'self' http://cdn.authgear.com; img-src http: https: data: 'self' http://cdn.authgear.com; object-src 'none'; base-uri 'none'; connect-src 'self' https://www.google-analytics.com ws://authgear.com wss://authgear.com; block-all-mixed-content; frame-ancestors http://authgearportal.com",
						},
					},
				},
				{
					AllowedFrameAncestorsFromEnv: config.AllowedFrameAncestors{"http://authgearportal.com"},
					OAuthConfig: &config.OAuthConfig{
						Clients: []config.OAuthClientConfig{
							{
								CustomUIURI: "http://customui.com",
							},
						},
					},
					AllowInlineScript:               false,
					AllowFrameAncestorsFromEnv:      false,
					AllowFrameAncestorsFromCustomUI: true,
					ExpectedHeaders: map[string][]string{
						"Content-Security-Policy": {
							"default-src 'self'; script-src 'strict-dynamic' 'nonce-' www.googletagmanager.com eu-assets.i.posthog.com challenges.cloudflare.com www.google.com https://browser.sentry-cdn.com 'self' http://cdn.authgear.com; frame-src www.googletagmanager.com challenges.cloudflare.com www.google.com 'self'; font-src cdnjs.cloudflare.com static2.sharepointonline.com fonts.googleapis.com fonts.gstatic.com 'self' http://cdn.authgear.com; style-src 'unsafe-inline' cdnjs.cloudflare.com www.googletagmanager.com fonts.googleapis.com 'self' http://cdn.authgear.com; img-src http: https: data: 'self' http://cdn.authgear.com; object-src 'none'; base-uri 'none'; connect-src 'self' https://www.google-analytics.com ws://authgear.com wss://authgear.com; block-all-mixed-content; frame-ancestors http://customui.com",
						},
					},
				},
				{
					AllowedFrameAncestorsFromEnv: config.AllowedFrameAncestors{"http://authgearportal.com"},
					OAuthConfig: &config.OAuthConfig{
						Clients: []config.OAuthClientConfig{
							{
								CustomUIURI: "http://customui.com",
							},
						},
					},
					AllowInlineScript:               false,
					AllowFrameAncestorsFromEnv:      false,
					AllowFrameAncestorsFromCustomUI: false,
					ExpectedHeaders: map[string][]string{
						"Content-Security-Policy": {
							"default-src 'self'; script-src 'strict-dynamic' 'nonce-' www.googletagmanager.com eu-assets.i.posthog.com challenges.cloudflare.com www.google.com https://browser.sentry-cdn.com 'self' http://cdn.authgear.com; frame-src www.googletagmanager.com challenges.cloudflare.com www.google.com 'self'; font-src cdnjs.cloudflare.com static2.sharepointonline.com fonts.googleapis.com fonts.gstatic.com 'self' http://cdn.authgear.com; style-src 'unsafe-inline' cdnjs.cloudflare.com www.googletagmanager.com fonts.googleapis.com 'self' http://cdn.authgear.com; img-src http: https: data: 'self' http://cdn.authgear.com; object-src 'none'; base-uri 'none'; connect-src 'self' https://www.google-analytics.com ws://authgear.com wss://authgear.com; block-all-mixed-content; frame-ancestors 'none'",
						},
						"X-Frame-Options": {"DENY"},
					},
				},
				{
					AllowedFrameAncestorsFromEnv:    config.AllowedFrameAncestors{},
					OAuthConfig:                     &config.OAuthConfig{},
					AllowInlineScript:               true,
					AllowFrameAncestorsFromEnv:      true,
					AllowFrameAncestorsFromCustomUI: true,
					ExpectedHeaders: map[string][]string{
						"Content-Security-Policy": {
							"default-src 'self'; script-src 'unsafe-inline' www.googletagmanager.com eu-assets.i.posthog.com challenges.cloudflare.com www.google.com https://browser.sentry-cdn.com 'self' http://cdn.authgear.com; frame-src www.googletagmanager.com challenges.cloudflare.com www.google.com 'self'; font-src cdnjs.cloudflare.com static2.sharepointonline.com fonts.googleapis.com fonts.gstatic.com 'self' http://cdn.authgear.com; style-src 'unsafe-inline' cdnjs.cloudflare.com www.googletagmanager.com fonts.googleapis.com 'self' http://cdn.authgear.com; img-src http: https: data: 'self' http://cdn.authgear.com; object-src 'none'; base-uri 'none'; connect-src 'self' https://www.google-analytics.com ws://authgear.com wss://authgear.com; block-all-mixed-content; frame-ancestors 'none'",
						},
						"X-Frame-Options": {"DENY"},
					},
				},
				{
					AllowedFrameAncestorsFromEnv:    config.AllowedFrameAncestors{},
					OAuthConfig:                     &config.OAuthConfig{},
					AllowInlineScript:               true,
					AllowFrameAncestorsFromEnv:      false,
					AllowFrameAncestorsFromCustomUI: false,
					ExpectedHeaders: map[string][]string{
						"Content-Security-Policy": {
							"default-src 'self'; script-src 'unsafe-inline' www.googletagmanager.com eu-assets.i.posthog.com challenges.cloudflare.com www.google.com https://browser.sentry-cdn.com 'self' http://cdn.authgear.com; frame-src www.googletagmanager.com challenges.cloudflare.com www.google.com 'self'; font-src cdnjs.cloudflare.com static2.sharepointonline.com fonts.googleapis.com fonts.gstatic.com 'self' http://cdn.authgear.com; style-src 'unsafe-inline' cdnjs.cloudflare.com www.googletagmanager.com fonts.googleapis.com 'self' http://cdn.authgear.com; img-src http: https: data: 'self' http://cdn.authgear.com; object-src 'none'; base-uri 'none'; connect-src 'self' https://www.google-analytics.com ws://authgear.com wss://authgear.com; block-all-mixed-content; frame-ancestors 'none'",
						},
						"X-Frame-Options": {"DENY"},
					},
				},
			}

			test := func(testcase TestCase, resultHeaders map[string][]string) {
				middleware := DynamicCSPMiddleware{
					Cookies:                         cookieManager,
					HTTPOrigin:                      "http://authgear.com",
					WebAppCDNHost:                   "http://cdn.authgear.com",
					AllowedFrameAncestorsFromEnv:    testcase.AllowedFrameAncestorsFromEnv,
					OAuthConfig:                     testcase.OAuthConfig,
					AllowInlineScript:               AllowInlineScript(testcase.AllowInlineScript),
					AllowFrameAncestorsFromEnv:      AllowFrameAncestorsFromEnv(testcase.AllowFrameAncestorsFromEnv),
					AllowFrameAncestorsFromCustomUI: AllowFrameAncestorsFromCustomUI(testcase.AllowFrameAncestorsFromCustomUI),
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
