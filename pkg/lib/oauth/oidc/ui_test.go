package oidc

import (
	"net/url"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/oauth/oauthsession"
	"github.com/authgear/authgear-server/pkg/lib/oauth/protocol"
)

type mockUIURLBuilderEndpoints struct{}

func (mockUIURLBuilderEndpoints) OAuthEntrypointURL() *url.URL {
	u, _ := url.Parse("https://authui.example.com/oauth_entrypoint")
	return u
}
func (mockUIURLBuilderEndpoints) SettingsChangePasswordURL() *url.URL     { return &url.URL{} }
func (mockUIURLBuilderEndpoints) SettingsDeleteAccountURL() *url.URL      { return &url.URL{} }
func (mockUIURLBuilderEndpoints) SettingsIdentityOAuthURL() *url.URL      { return &url.URL{} }
func (mockUIURLBuilderEndpoints) SettingsAddLoginIDEmail(string) *url.URL { return &url.URL{} }
func (mockUIURLBuilderEndpoints) SettingsAddLoginIDPhone(string) *url.URL { return &url.URL{} }
func (mockUIURLBuilderEndpoints) SettingsAddLoginIDUsername(string) *url.URL {
	return &url.URL{}
}
func (mockUIURLBuilderEndpoints) SettingsEditLoginIDEmail(string) *url.URL { return &url.URL{} }
func (mockUIURLBuilderEndpoints) SettingsEditLoginIDPhone(string) *url.URL { return &url.URL{} }
func (mockUIURLBuilderEndpoints) SettingsEditLoginIDUsername(string) *url.URL {
	return &url.URL{}
}

func TestUIURLBuilderBuildAuthenticationURL(t *testing.T) {
	Convey("UIURLBuilder.BuildAuthenticationURL", t, func() {
		b := &UIURLBuilder{
			Endpoints:      mockUIURLBuilderEndpoints{},
			IdentityConfig: &config.IdentityConfig{},
		}
		e := &oauthsession.Entry{ID: "oauthsession_abc123"}
		req := protocol.AuthorizationRequest{
			"client_id":     "e2e",
			"login_hint":    "https://example.com?login_hint=login_id%3Aemail%3Auser%40example.com",
			"id_token_hint": "some.jwt.token",
		}

		Convey("forwards login_hint and id_token_hint to a Custom UI", func() {
			client := &config.OAuthClientConfig{
				ClientID:    "e2e",
				CustomUIURI: "https://ui.example.com/auth",
			}
			endpoint, err := b.BuildAuthenticationURL(client, req, e)
			So(err, ShouldBeNil)
			So(endpoint.Query().Get("login_hint"), ShouldEqual, req["login_hint"])
			So(endpoint.Query().Get("id_token_hint"), ShouldEqual, req["id_token_hint"])
		})

		Convey("does not add login_hint/id_token_hint when the request has neither", func() {
			client := &config.OAuthClientConfig{
				ClientID:    "e2e",
				CustomUIURI: "https://ui.example.com/auth",
			}
			bareReq := protocol.AuthorizationRequest{"client_id": "e2e"}
			endpoint, err := b.BuildAuthenticationURL(client, bareReq, e)
			So(err, ShouldBeNil)
			So(endpoint.Query().Has("login_hint"), ShouldBeFalse)
			So(endpoint.Query().Has("id_token_hint"), ShouldBeFalse)
		})

		Convey("does not forward login_hint/id_token_hint to the built-in AuthUI", func() {
			client := &config.OAuthClientConfig{ClientID: "e2e"}
			endpoint, err := b.BuildAuthenticationURL(client, req, e)
			So(err, ShouldBeNil)
			So(endpoint.Query().Has("login_hint"), ShouldBeFalse)
			So(endpoint.Query().Has("id_token_hint"), ShouldBeFalse)
		})
	})
}
