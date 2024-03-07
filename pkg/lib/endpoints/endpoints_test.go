package endpoints

import (
	"net/url"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestEndpoints(t *testing.T) {
	Convey("Endpoints", t, func() {
		endpoints := Endpoints{
			HTTPProto: "https",
			HTTPHost:  "example.com",
		}

		So(endpoints.Origin(), ShouldResemble, &url.URL{
			Scheme: "https",
			Host:   "example.com",
		})

		So(endpoints.AuthorizeEndpointURL().String(), ShouldEqual, "https://example.com/oauth2/authorize")
		So(endpoints.ConsentEndpointURL().String(), ShouldEqual, "https://example.com/oauth2/consent")
		So(endpoints.TokenEndpointURL().String(), ShouldEqual, "https://example.com/oauth2/token")
		So(endpoints.RevokeEndpointURL().String(), ShouldEqual, "https://example.com/oauth2/revoke")
		So(endpoints.JWKSEndpointURL().String(), ShouldEqual, "https://example.com/oauth2/jwks")
		So(endpoints.UserInfoEndpointURL().String(), ShouldEqual, "https://example.com/oauth2/userinfo")
		So(endpoints.EndSessionEndpointURL().String(), ShouldEqual, "https://example.com/oauth2/end_session")
		So(endpoints.OAuthEntrypointURL().String(), ShouldEqual, "https://example.com/_internals/oauth_entrypoint")
		So(endpoints.LoginEndpointURL().String(), ShouldEqual, "https://example.com/login")
		So(endpoints.SignupEndpointURL().String(), ShouldEqual, "https://example.com/signup")
		So(endpoints.PromoteUserEndpointURL().String(), ShouldEqual, "https://example.com/flows/promote_user")
		So(endpoints.LogoutEndpointURL().String(), ShouldEqual, "https://example.com/logout")
		So(endpoints.SettingsEndpointURL().String(), ShouldEqual, "https://example.com/settings")
		So(endpoints.SSOCallbackEndpointURL().String(), ShouldEqual, "https://example.com/sso/oauth2/callback")
		So(endpoints.SSOCallbackURL("google").String(), ShouldEqual, "https://example.com/sso/oauth2/callback/google")
	})
}
