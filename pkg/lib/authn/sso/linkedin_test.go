package sso

import (
	"testing"

	"github.com/authgear/authgear-server/pkg/lib/config"
	. "github.com/smartystreets/goconvey/convey"
)

func TestLinkedInImpl(t *testing.T) {
	Convey("LinkedInImpl", t, func() {
		g := &LinkedInImpl{
			ProviderConfig: config.OAuthSSOProviderConfig{
				ClientID: "client_id",
				Type:     config.OAuthSSOProviderTypeLinkedIn,
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
		So(u, ShouldEqual, "https://www.linkedin.com/oauth/v2/authorization?client_id=client_id&redirect_uri=https%3A%2F%2Flocalhost%2F&response_type=code&scope=r_liteprofile+r_emailaddress&state=state")
	})
}
