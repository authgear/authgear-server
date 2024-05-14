package github

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/oauthrelyingparty/pkg/api/oauthrelyingparty"
)

func TestGithub(t *testing.T) {
	Convey("Github", t, func() {
		deps := oauthrelyingparty.Dependencies{
			ProviderConfig: oauthrelyingparty.ProviderConfig{
				"client_id": "client_id",
				"type":      Type,
			},
		}
		g := Github{}

		u, err := g.GetAuthorizationURL(deps, oauthrelyingparty.GetAuthorizationURLOptions{
			RedirectURI: "https://localhost/",
			Nonce:       "nonce",
			State:       "state",
			Prompt:      []string{"login"},
		})
		So(err, ShouldBeNil)
		So(u, ShouldEqual, "https://github.com/login/oauth/authorize?client_id=client_id&redirect_uri=https%3A%2F%2Flocalhost%2F&scope=read%3Auser+user%3Aemail&state=state")
	})
}
