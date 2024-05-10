package sso

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/api/oauthrelyingparty"
	"github.com/authgear/authgear-server/pkg/lib/oauthrelyingparty/github"
)

func TestGithubImpl(t *testing.T) {
	Convey("GithubImpl", t, func() {
		deps := oauthrelyingparty.Dependencies{
			ProviderConfig: oauthrelyingparty.ProviderConfig{
				"client_id": "client_id",
				"type":      github.Type,
			},
		}
		g := &GithubImpl{}

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
