package sso

import (
	"net/http"
	"testing"

	"github.com/authgear/authgear-server/pkg/lib/config"
	. "github.com/smartystreets/goconvey/convey"
	"gopkg.in/h2non/gock.v1"
)

func TestADFSImpl(t *testing.T) {
	Convey("ADFSImpl", t, func() {
		gock.InterceptClient(http.DefaultClient)
		defer gock.Off()

		g := &ADFSImpl{
			RedirectURL: mockRedirectURLProvider{},
			ProviderConfig: config.OAuthSSOProviderConfig{
				ClientID:                  "client_id",
				Type:                      config.OAuthSSOProviderTypeADFS,
				DiscoveryDocumentEndpoint: "https://localhost/.well-known/openid-configuration",
			},
		}

		gock.New("https://localhost/.well-known/openid-configuration").
			Reply(200).
			BodyString(`
{
  "authorization_endpoint": "https://localhost/authorize"
}
			`)
		defer func() { gock.Flush() }()

		u, err := g.GetAuthURL(GetAuthURLParam{
			Nonce:  "nonce",
			State:  "state",
			Prompt: []string{"login"},
		})
		So(err, ShouldBeNil)
		So(u, ShouldEqual, "https://localhost/authorize?client_id=client_id&nonce=nonce&prompt=login&redirect_uri=https%3A%2F%2Flocalhost%2F&response_mode=form_post&response_type=code&scope=openid+profile+email&state=state")
	})
}
