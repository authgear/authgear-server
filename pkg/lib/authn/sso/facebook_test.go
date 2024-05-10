package sso

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/api/oauthrelyingparty"
	"github.com/authgear/authgear-server/pkg/lib/oauthrelyingparty/facebook"
)

func TestFacebookImpl(t *testing.T) {
	Convey("FacebookImpl", t, func() {
		g := &FacebookImpl{
			ProviderConfig: oauthrelyingparty.ProviderConfig{
				"client_id": "client_id",
				"type":      facebook.Type,
			},
			HTTPClient: OAuthHTTPClient{},
		}

		u, err := g.GetAuthorizationURL(oauthrelyingparty.GetAuthorizationURLOptions{
			RedirectURI: "https://localhost/",
			Nonce:       "nonce",
			State:       "state",
			Prompt:      []string{"login"},
		})
		So(err, ShouldBeNil)
		So(u, ShouldEqual, "https://www.facebook.com/v11.0/dialog/oauth?client_id=client_id&redirect_uri=https%3A%2F%2Flocalhost%2F&response_type=code&scope=email+public_profile&state=state")
	})
}
