package sso

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/api/oauthrelyingparty"
	"github.com/authgear/authgear-server/pkg/lib/oauthrelyingparty/linkedin"
)

func TestLinkedInImpl(t *testing.T) {
	Convey("LinkedInImpl", t, func() {
		deps := oauthrelyingparty.Dependencies{
			ProviderConfig: oauthrelyingparty.ProviderConfig{
				"client_id": "client_id",
				"type":      linkedin.Type,
			},
		}
		g := &LinkedInImpl{}

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
