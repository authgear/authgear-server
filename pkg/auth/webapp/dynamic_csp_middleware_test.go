package webapp

import (
	"net/http"
	"net/http/httptest"
	"testing"

	gomock "github.com/golang/mock/gomock"
	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/httputil"
)

func TestDynamicCSPMiddleware(t *testing.T) {
	Convey("DynamicCSPMiddleware", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		cookieManager := NewMockCookieManager(ctrl)
		cookieManager.EXPECT().GetCookie(gomock.Any(), httputil.CSPNonceCookieDef).Return(&http.Cookie{}, nil).AnyTimes()

		type TestCase struct {
			AllowedFrameAncestorsFromEnv    config.AllowedFrameAncestors
			OAuthConfig                     *config.OAuthConfig
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
					AllowFrameAncestorsFromEnv:      true,
					AllowFrameAncestorsFromCustomUI: true,
					ExpectedHeaders: map[string][]string{
						"Content-Security-Policy": {
							"script-src 'self' https: 'nonce-' 'strict-dynamic'; object-src 'none'; base-uri 'none'; frame-ancestors http://authgearportal.com http://customui.com",
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
					AllowFrameAncestorsFromEnv:      true,
					AllowFrameAncestorsFromCustomUI: false,
					ExpectedHeaders: map[string][]string{
						"Content-Security-Policy": {
							"script-src 'self' https: 'nonce-' 'strict-dynamic'; object-src 'none'; base-uri 'none'; frame-ancestors http://authgearportal.com",
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
					AllowFrameAncestorsFromEnv:      false,
					AllowFrameAncestorsFromCustomUI: true,
					ExpectedHeaders: map[string][]string{
						"Content-Security-Policy": {
							"script-src 'self' https: 'nonce-' 'strict-dynamic'; object-src 'none'; base-uri 'none'; frame-ancestors http://customui.com",
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
					AllowFrameAncestorsFromEnv:      false,
					AllowFrameAncestorsFromCustomUI: false,
					ExpectedHeaders: map[string][]string{
						"Content-Security-Policy": {
							"script-src 'self' https: 'nonce-' 'strict-dynamic'; object-src 'none'; base-uri 'none'; frame-ancestors 'none'",
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
					AllowFrameAncestorsFromEnv:      true,
					AllowFrameAncestorsFromCustomUI: true,
					ExpectedHeaders: map[string][]string{
						"Content-Security-Policy": {
							"script-src 'self' https: 'nonce-' 'strict-dynamic'; object-src 'none'; base-uri 'none'; frame-ancestors http://authgearportal.com http://customui.com",
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
					AllowFrameAncestorsFromEnv:      true,
					AllowFrameAncestorsFromCustomUI: false,
					ExpectedHeaders: map[string][]string{
						"Content-Security-Policy": {
							"script-src 'self' https: 'nonce-' 'strict-dynamic'; object-src 'none'; base-uri 'none'; frame-ancestors http://authgearportal.com",
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
					AllowFrameAncestorsFromEnv:      false,
					AllowFrameAncestorsFromCustomUI: true,
					ExpectedHeaders: map[string][]string{
						"Content-Security-Policy": {
							"script-src 'self' https: 'nonce-' 'strict-dynamic'; object-src 'none'; base-uri 'none'; frame-ancestors http://customui.com",
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
					AllowFrameAncestorsFromEnv:      false,
					AllowFrameAncestorsFromCustomUI: false,
					ExpectedHeaders: map[string][]string{
						"Content-Security-Policy": {
							"script-src 'self' https: 'nonce-' 'strict-dynamic'; object-src 'none'; base-uri 'none'; frame-ancestors 'none'",
						},
						"X-Frame-Options": {"DENY"},
					},
				},
				{
					AllowedFrameAncestorsFromEnv:    config.AllowedFrameAncestors{},
					OAuthConfig:                     &config.OAuthConfig{},
					AllowFrameAncestorsFromEnv:      true,
					AllowFrameAncestorsFromCustomUI: true,
					ExpectedHeaders: map[string][]string{
						"Content-Security-Policy": {
							"script-src 'self' https: 'nonce-' 'strict-dynamic'; object-src 'none'; base-uri 'none'; frame-ancestors 'none'",
						},
						"X-Frame-Options": {"DENY"},
					},
				},
				{
					AllowedFrameAncestorsFromEnv:    config.AllowedFrameAncestors{},
					OAuthConfig:                     &config.OAuthConfig{},
					AllowFrameAncestorsFromEnv:      false,
					AllowFrameAncestorsFromCustomUI: false,
					ExpectedHeaders: map[string][]string{
						"Content-Security-Policy": {
							"script-src 'self' https: 'nonce-' 'strict-dynamic'; object-src 'none'; base-uri 'none'; frame-ancestors 'none'",
						},
						"X-Frame-Options": {"DENY"},
					},
				},
			}

			test := func(testcase TestCase, resultHeaders map[string][]string) {
				middleware := DynamicCSPMiddleware{
					Cookies:                         cookieManager,
					AllowedFrameAncestorsFromEnv:    testcase.AllowedFrameAncestorsFromEnv,
					OAuthConfig:                     testcase.OAuthConfig,
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
