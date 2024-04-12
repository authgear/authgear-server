package sso

import (
	"testing"

	"github.com/authgear/authgear-server/pkg/lib/config"
	. "github.com/smartystreets/goconvey/convey"
)

func TestAppleImpl(t *testing.T) {
	Convey("AppleImpl", t, func() {
		g := &AppleImpl{
			ProviderConfig: config.OAuthSSOProviderConfig{
				ClientID: "client_id",
				Type:     config.OAuthSSOProviderTypeApple,
			},
			HTTPClient: OAuthHTTPClient{},
		}

		u, err := g.GetAuthURL(GetAuthURLParam{
			RedirectURI:  "https://localhost/",
			ResponseMode: ResponseModeFormPost,
			Nonce:        "nonce",
			State:        "state",
			Prompt:       []string{"login"},
		})
		So(err, ShouldBeNil)
		So(u, ShouldEqual, "https://appleid.apple.com/auth/authorize?client_id=client_id&nonce=nonce&redirect_uri=https%3A%2F%2Flocalhost%2F&response_mode=form_post&response_type=code&scope=name+email&state=state")
	})
}
