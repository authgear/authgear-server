package handler

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/oauth/protocol"
	"github.com/authgear/authgear-server/pkg/util/httputil"
)

type mockOAuthRequestImpl struct {
	redirectURI string
}

func (o *mockOAuthRequestImpl) ClientID() string {
	return ""
}

func (o *mockOAuthRequestImpl) RedirectURI() string {
	return o.redirectURI
}

func TestParseRedirectURI(t *testing.T) {
	clientConfig := &config.OAuthClientConfig{
		RedirectURIs: []string{
			"http://app.example.com/handle_auth",
			"com.example.myapp://host/path",
		},
		CustomUIURI: "http://authui.example.com/auth",
	}

	httpOrigin := httputil.HTTPOrigin("http://auth.example.com")
	httpProto := httputil.HTTPProto("http")
	whitelistedDomains := []string{
		"auth.example2.com",
		"auth.example3.com",
	}

	Convey("parseRedirectURI", t, func() {
		Convey("should use default redirect uri", func() {
			u, err := parseRedirectURI(&config.OAuthClientConfig{
				RedirectURIs: []string{
					"http://app.example.com/handle_auth",
				},
			}, httpProto, httpOrigin, whitelistedDomains, []string{}, &mockOAuthRequestImpl{})

			So(u.String(), ShouldResemble, "http://app.example.com/handle_auth")
			So(err, ShouldBeNil)
		})

		Convey("should allow allowlisted redirect uri", func() {
			u, err := parseRedirectURI(clientConfig, httpProto, httpOrigin, whitelistedDomains, []string{}, &mockOAuthRequestImpl{
				"com.example.myapp://host/path",
			})

			So(u.String(), ShouldResemble, "com.example.myapp://host/path")
			So(err, ShouldBeNil)
		})

		Convey("should exact match", func() {
			_, err := parseRedirectURI(clientConfig, httpProto, httpOrigin, whitelistedDomains, []string{}, &mockOAuthRequestImpl{
				"http://app.example.com/handle_auth/",
			})

			So(err, ShouldResemble, protocol.NewErrorResponse("invalid_request", "redirect URI is not allowed"))
		})

		Convey("should allow URIs at same origin as the authgear server", func() {
			u, err := parseRedirectURI(clientConfig, httpProto, httpOrigin, whitelistedDomains, []string{}, &mockOAuthRequestImpl{
				"http://auth.example.com/settings",
			})

			So(u.String(), ShouldResemble, "http://auth.example.com/settings")
			So(err, ShouldBeNil)
		})

		Convey("should allow URIs at same origin as the custom ui uri", func() {
			u, err := parseRedirectURI(clientConfig, httpProto, httpOrigin, whitelistedDomains, []string{}, &mockOAuthRequestImpl{
				"http://authui.example.com/auth/complete",
			})

			So(u.String(), ShouldResemble, "http://authui.example.com/auth/complete")
			So(err, ShouldBeNil)
		})

		Convey("should allow URIs with domain in whitelist in same protocol", func() {
			u1, err := parseRedirectURI(clientConfig, httpProto, httpOrigin, whitelistedDomains, []string{}, &mockOAuthRequestImpl{
				"http://auth.example2.com/auth/complete",
			})

			So(u1.String(), ShouldResemble, "http://auth.example2.com/auth/complete")
			So(err, ShouldBeNil)

			u2, err := parseRedirectURI(clientConfig, httpProto, httpOrigin, whitelistedDomains, []string{}, &mockOAuthRequestImpl{
				"http://auth.example3.com/auth/complete",
			})

			So(u2.String(), ShouldResemble, "http://auth.example3.com/auth/complete")
			So(err, ShouldBeNil)
		})

		Convey("should reject URIs not in the allowlist", func() {
			_, err := parseRedirectURI(clientConfig, httpProto, httpOrigin, whitelistedDomains, []string{}, &mockOAuthRequestImpl{
				"http://unknown.com",
			})

			So(err, ShouldResemble, protocol.NewErrorResponse("invalid_request", "redirect URI is not allowed"))
		})

		Convey("should allow any port for a registered loopback redirect uri", func() {
			// RFC 8252 §7.3: the authorization server MUST allow any port to be
			// specified at the time of the request for loopback redirect URIs.
			loopbackClientConfig := &config.OAuthClientConfig{
				RedirectURIs: []string{
					"http://127.0.0.1/callback",
					"http://[::1]/callback",
					"http://localhost:1234/callback",
					"https://app.example.com/callback",
				},
			}

			Convey("should allow an ephemeral port on the IPv4 loopback", func() {
				u, err := parseRedirectURI(loopbackClientConfig, httpProto, httpOrigin, []string{}, []string{}, &mockOAuthRequestImpl{
					"http://127.0.0.1:53412/callback",
				})

				So(u.String(), ShouldResemble, "http://127.0.0.1:53412/callback")
				So(err, ShouldBeNil)
			})

			Convey("should allow an ephemeral port on the IPv6 loopback", func() {
				u, err := parseRedirectURI(loopbackClientConfig, httpProto, httpOrigin, []string{}, []string{}, &mockOAuthRequestImpl{
					"http://[::1]:61023/callback",
				})

				So(u.String(), ShouldResemble, "http://[::1]:61023/callback")
				So(err, ShouldBeNil)
			})

			Convey("should allow a different port than the registered one", func() {
				u, err := parseRedirectURI(loopbackClientConfig, httpProto, httpOrigin, []string{}, []string{}, &mockOAuthRequestImpl{
					"http://localhost:53412/callback",
				})

				So(u.String(), ShouldResemble, "http://localhost:53412/callback")
				So(err, ShouldBeNil)
			})

			Convey("should still allow the registered loopback uri itself", func() {
				u, err := parseRedirectURI(loopbackClientConfig, httpProto, httpOrigin, []string{}, []string{}, &mockOAuthRequestImpl{
					"http://127.0.0.1/callback",
				})

				So(u.String(), ShouldResemble, "http://127.0.0.1/callback")
				So(err, ShouldBeNil)
			})

			// This case has to use a config of its own. Every loopback entry in
			// loopbackClientConfig is registered with the /callback path, so
			// against that config no incoming URI can differ from all of them
			// by host alone -- a rejection there would prove nothing about the
			// host comparison, since the path would reject it anyway.
			Convey("should reject a loopback host that is not registered, even on a registered path", func() {
				_, err := parseRedirectURI(&config.OAuthClientConfig{
					RedirectURIs: []string{"http://127.0.0.1/callback"},
				}, httpProto, httpOrigin, []string{}, []string{}, &mockOAuthRequestImpl{
					"http://localhost:8000/callback",
				})

				So(err, ShouldResemble, protocol.NewErrorResponse("invalid_request", "redirect URI is not allowed"))
			})

			Convey("should reject a registered loopback host on an unregistered path", func() {
				_, err := parseRedirectURI(&config.OAuthClientConfig{
					RedirectURIs: []string{"http://127.0.0.1/callback"},
				}, httpProto, httpOrigin, []string{}, []string{}, &mockOAuthRequestImpl{
					"http://127.0.0.1:53412/other",
				})

				So(err, ShouldResemble, protocol.NewErrorResponse("invalid_request", "redirect URI is not allowed"))
			})

			Convey("should reject a different path", func() {
				_, err := parseRedirectURI(loopbackClientConfig, httpProto, httpOrigin, []string{}, []string{}, &mockOAuthRequestImpl{
					"http://127.0.0.1:53412/other",
				})

				So(err, ShouldResemble, protocol.NewErrorResponse("invalid_request", "redirect URI is not allowed"))
			})

			Convey("should reject a different query", func() {
				_, err := parseRedirectURI(loopbackClientConfig, httpProto, httpOrigin, []string{}, []string{}, &mockOAuthRequestImpl{
					"http://127.0.0.1:53412/callback?evil=1",
				})

				So(err, ShouldResemble, protocol.NewErrorResponse("invalid_request", "redirect URI is not allowed"))
			})

			Convey("should reject the https scheme", func() {
				_, err := parseRedirectURI(loopbackClientConfig, httpProto, httpOrigin, []string{}, []string{}, &mockOAuthRequestImpl{
					"https://127.0.0.1:53412/callback",
				})

				So(err, ShouldResemble, protocol.NewErrorResponse("invalid_request", "redirect URI is not allowed"))
			})

			Convey("should reject a host that merely looks like the loopback", func() {
				_, err := parseRedirectURI(loopbackClientConfig, httpProto, httpOrigin, []string{}, []string{}, &mockOAuthRequestImpl{
					"http://127.0.0.1.evil.com:53412/callback",
				})

				So(err, ShouldResemble, protocol.NewErrorResponse("invalid_request", "redirect URI is not allowed"))
			})

			Convey("should not ignore the port of a non-loopback redirect uri", func() {
				_, err := parseRedirectURI(loopbackClientConfig, httpProto, httpOrigin, []string{}, []string{}, &mockOAuthRequestImpl{
					"https://app.example.com:9000/callback",
				})

				So(err, ShouldResemble, protocol.NewErrorResponse("invalid_request", "redirect URI is not allowed"))
			})

			Convey("should not ignore the port of a non-loopback redirect uri over http", func() {
				_, err := parseRedirectURI(clientConfig, httpProto, httpOrigin, []string{}, []string{}, &mockOAuthRequestImpl{
					"http://app.example.com:9000/handle_auth",
				})

				So(err, ShouldResemble, protocol.NewErrorResponse("invalid_request", "redirect URI is not allowed"))
			})
		})

		Convey("should allow origins in allowlist", func() {
			u, err := parseRedirectURI(clientConfig, httpProto, httpOrigin, []string{}, []string{
				"http://anotheroriginexample.com",
			}, &mockOAuthRequestImpl{
				"http://anotheroriginexample.com/?q=test#test",
			})

			So(u.String(), ShouldResemble, "http://anotheroriginexample.com/?q=test#test")
			So(err, ShouldBeNil)
		})
	})
}
