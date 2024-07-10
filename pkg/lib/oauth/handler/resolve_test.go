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
