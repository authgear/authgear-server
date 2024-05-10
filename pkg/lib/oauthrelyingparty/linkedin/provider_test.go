package linkedin

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/api/oauthrelyingparty"
)

func TestLinkedin(t *testing.T) {
	Convey("Linkedin", t, func() {
		deps := oauthrelyingparty.Dependencies{
			ProviderConfig: oauthrelyingparty.ProviderConfig{
				"client_id": "client_id",
				"type":      Type,
			},
		}
		g := Linkedin{}

		u, err := g.GetAuthorizationURL(deps, oauthrelyingparty.GetAuthorizationURLOptions{
			RedirectURI: "https://localhost/",
			Nonce:       "nonce",
			State:       "state",
			Prompt:      []string{"login"},
		})
		So(err, ShouldBeNil)
		So(u, ShouldEqual, "https://www.linkedin.com/oauth/v2/authorization?client_id=client_id&redirect_uri=https%3A%2F%2Flocalhost%2F&response_type=code&scope=r_liteprofile+r_emailaddress&state=state")
	})
}
