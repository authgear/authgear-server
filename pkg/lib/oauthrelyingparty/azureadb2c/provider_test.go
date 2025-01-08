package azureadb2c

import (
	"context"
	"net/http"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"gopkg.in/h2non/gock.v1"

	"github.com/authgear/oauthrelyingparty/pkg/api/oauthrelyingparty"
)

func TestAzureadb2cImpl(t *testing.T) {
	Convey("Azureadb2cImpl.GetAuthorizationURL", t, func() {
		client := &http.Client{}
		gock.InterceptClient(client)
		defer gock.Off()

		test := func(providerConfig oauthrelyingparty.ProviderConfig, expected string) {
			deps := oauthrelyingparty.Dependencies{
				ProviderConfig: providerConfig,
				HTTPClient:     client,
			}

			g := AzureADB2C{}

			gock.New("https://tenant.b2clogin.com/tenant.onmicrosoft.com/policy/v2.0/.well-known/openid-configuration").
				Reply(200).
				BodyString(`
{
  "authorization_endpoint": "https://localhost/authorize"
}
			`)
			defer func() { gock.Flush() }()

			ctx := context.Background()
			u, err := g.GetAuthorizationURL(ctx, deps, oauthrelyingparty.GetAuthorizationURLOptions{
				RedirectURI:  "https://localhost/",
				ResponseMode: oauthrelyingparty.ResponseModeFormPost,
				Nonce:        "nonce",
				State:        "state",
				Prompt:       []string{"login"},
			})
			So(err, ShouldBeNil)
			So(u, ShouldEqual, expected)
		}

		Convey("without domain_hint", func() {
			test(oauthrelyingparty.ProviderConfig{
				"client_id": "client_id",
				"type":      Type,
				"tenant":    "tenant",
				"policy":    "policy",
			}, "https://localhost/authorize?client_id=client_id&nonce=nonce&prompt=login&redirect_uri=https%3A%2F%2Flocalhost%2F&response_mode=form_post&response_type=code&scope=openid&state=state")
		})

		Convey("with domain_hint", func() {
			test(oauthrelyingparty.ProviderConfig{
				"client_id":   "client_id",
				"type":        Type,
				"tenant":      "tenant",
				"policy":      "policy",
				"domain_hint": "google.com",
			}, "https://localhost/authorize?client_id=client_id&domain_hint=google.com&nonce=nonce&prompt=login&redirect_uri=https%3A%2F%2Flocalhost%2F&response_mode=form_post&response_type=code&scope=openid&state=state")
		})
	})

}
