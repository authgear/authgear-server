package linkedin

import (
	"net/http"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"gopkg.in/h2non/gock.v1"

	"github.com/authgear/oauthrelyingparty/pkg/api/oauthrelyingparty"
)

func TestLinkedin(t *testing.T) {
	Convey("Linkedin", t, func() {
		client := &http.Client{}
		gock.InterceptClient(client)
		defer gock.Off()

		deps := oauthrelyingparty.Dependencies{
			ProviderConfig: oauthrelyingparty.ProviderConfig{
				"client_id": "client_id",
				"type":      Type,
			},
			HTTPClient: client,
		}
		g := Linkedin{}

		gock.New(linkedinOIDCDiscoveryDocumentURL).Reply(200).BodyString(`
{
  "issuer": "https://www.linkedin.com/oauth",
  "authorization_endpoint": "https://www.linkedin.com/oauth/v2/authorization",
  "token_endpoint": "https://www.linkedin.com/oauth/v2/accessToken",
  "userinfo_endpoint": "https://api.linkedin.com/v2/userinfo",
  "jwks_uri": "https://www.linkedin.com/oauth/openid/jwks",
  "response_types_supported": [
    "code"
  ],
  "subject_types_supported": [
    "pairwise"
  ],
  "id_token_signing_alg_values_supported": [
    "RS256"
  ],
  "scopes_supported": [
    "openid",
    "profile",
    "email"
  ],
  "claims_supported": [
    "iss",
    "aud",
    "iat",
    "exp",
    "sub",
    "name",
    "given_name",
    "family_name",
    "picture",
    "email",
    "email_verified",
    "locale"
  ]
}
		`)
		defer func() { gock.Flush() }()

		u, err := g.GetAuthorizationURL(deps, oauthrelyingparty.GetAuthorizationURLOptions{
			RedirectURI: "https://localhost/",
			State:       "state",
			Prompt:      []string{"login"},
		})
		So(err, ShouldBeNil)
		So(u, ShouldEqual, "https://www.linkedin.com/oauth/v2/authorization?client_id=client_id&redirect_uri=https%3A%2F%2Flocalhost%2F&response_type=code&scope=openid+profile+email&state=state")
	})
}
