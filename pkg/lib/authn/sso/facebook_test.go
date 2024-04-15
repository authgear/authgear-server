package sso

import (
	"testing"

	"github.com/authgear/authgear-server/pkg/lib/config"
	. "github.com/smartystreets/goconvey/convey"
)

func TestFacebookImpl(t *testing.T) {
	Convey("FacebookImpl", t, func() {
		g := &FacebookImpl{
			ProviderConfig: config.OAuthSSOProviderConfig{
				ClientID: "client_id",
				Type:     config.OAuthSSOProviderTypeFacebook,
			},
			HTTPClient: OAuthHTTPClient{},
		}

		u, err := g.GetAuthURL(GetAuthURLParam{
			RedirectURI: "https://localhost/",
			Nonce:       "nonce",
			State:       "state",
			Prompt:      []string{"login"},
		})
		So(err, ShouldBeNil)
		So(u, ShouldEqual, "https://www.facebook.com/v11.0/dialog/oauth?client_id=client_id&redirect_uri=https%3A%2F%2Flocalhost%2F&response_type=code&scope=email+public_profile&state=state")
	})
}
