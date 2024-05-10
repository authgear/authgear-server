package sso

import (
	"net/http"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"gopkg.in/h2non/gock.v1"

	"github.com/authgear/authgear-server/pkg/api/oauthrelyingparty"
	"github.com/authgear/authgear-server/pkg/lib/oauthrelyingparty/azureadb2c"
)

func TestAzureadb2cImpl(t *testing.T) {
	Convey("Azureadb2cImpl", t, func() {
		client := OAuthHTTPClient{&http.Client{}}
		gock.InterceptClient(client.Client)
		defer gock.Off()

		g := &Azureadb2cImpl{
			ProviderConfig: oauthrelyingparty.ProviderConfig{
				"client_id": "client_id",
				"type":      azureadb2c.Type,
				"tenant":    "tenant",
				"policy":    "policy",
			},
			HTTPClient: client,
		}

		gock.New("https://tenant.b2clogin.com/tenant.onmicrosoft.com/policy/v2.0/.well-known/openid-configuration").
			Reply(200).
			BodyString(`
{
  "authorization_endpoint": "https://localhost/authorize"
}
			`)
		defer func() { gock.Flush() }()

		u, err := g.GetAuthorizationURL(oauthrelyingparty.GetAuthorizationURLOptions{
			RedirectURI:  "https://localhost/",
			ResponseMode: oauthrelyingparty.ResponseModeFormPost,
			Nonce:        "nonce",
			State:        "state",
			Prompt:       []string{"login"},
		})
		So(err, ShouldBeNil)
		So(u, ShouldEqual, "https://localhost/authorize?client_id=client_id&nonce=nonce&prompt=login&redirect_uri=https%3A%2F%2Flocalhost%2F&response_mode=form_post&response_type=code&scope=openid&state=state")
	})
}
