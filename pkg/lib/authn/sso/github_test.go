package sso

import (
	"net/http"
	"testing"

	"github.com/authgear/authgear-server/pkg/lib/config"
	. "github.com/smartystreets/goconvey/convey"
)

func TestGithubImpl(t *testing.T) {
	Convey("GithubImpl", t, func() {
		g := &GithubImpl{
			ProviderConfig: config.OAuthSSOProviderConfig{
				ClientID: "client_id",
				Type:     config.OAuthSSOProviderTypeGithub,
			},
			HTTPClient: http.DefaultClient,
		}

		u, err := g.GetAuthURL(GetAuthURLParam{
			RedirectURI: "https://localhost/",
			Nonce:       "nonce",
			State:       "state",
			Prompt:      []string{"login"},
		})
		So(err, ShouldBeNil)
		So(u, ShouldEqual, "https://github.com/login/oauth/authorize?client_id=client_id&redirect_uri=https%3A%2F%2Flocalhost%2F&scope=read%3Auser+user%3Aemail&state=state")
	})
}
