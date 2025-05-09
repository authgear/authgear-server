package oauthrelyingpartyutil

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestOIDCDiscoveryDocumentWithRewrittenEndpoints(t *testing.T) {
	Convey("OIDCDiscoveryDocument.WithRewrittenEndpoints", t, func() {
		Convey("does not replace anything if no match", func() {
			So((&OIDCDiscoveryDocument{
				Issuer:                "https://accounts.myapp.com",
				AuthorizationEndpoint: "https://accounts.myapp.com/oauth2/authorize",
				TokenEndpoint:         "https://accounts.myapp.com/oauth2/token",
				UserInfoEndpoint:      "https://accounts.myapp.com/oauth2/userinfo",
				JWKSUri:               "https://accounts.myapp.com/oauth2/jwks",
			}).WithRewrittenEndpoints("https://a.myapp.com", "https://b.myapp.com"), ShouldResemble, &OIDCDiscoveryDocument{
				Issuer:                "https://accounts.myapp.com",
				AuthorizationEndpoint: "https://accounts.myapp.com/oauth2/authorize",
				TokenEndpoint:         "https://accounts.myapp.com/oauth2/token",
				UserInfoEndpoint:      "https://accounts.myapp.com/oauth2/userinfo",
				JWKSUri:               "https://accounts.myapp.com/oauth2/jwks",
			})
		})

		Convey("replace all matches", func() {
			So((&OIDCDiscoveryDocument{
				Issuer:                "https://accounts.myapp.com",
				AuthorizationEndpoint: "https://accounts.myapp.com/oauth2/authorize",
				TokenEndpoint:         "https://accounts.myapp.com/oauth2/token",
				UserInfoEndpoint:      "https://accounts.myapp.com/oauth2/userinfo",
				JWKSUri:               "https://accounts.myapp.com/oauth2/jwks",
			}).WithRewrittenEndpoints("https://accounts.myapp.com", "http://accounts.projects.authgear"), ShouldResemble, &OIDCDiscoveryDocument{
				Issuer:                "https://accounts.myapp.com",
				AuthorizationEndpoint: "http://accounts.projects.authgear/oauth2/authorize",
				TokenEndpoint:         "http://accounts.projects.authgear/oauth2/token",
				UserInfoEndpoint:      "http://accounts.projects.authgear/oauth2/userinfo",
				JWKSUri:               "http://accounts.projects.authgear/oauth2/jwks",
			})
		})
	})
}
