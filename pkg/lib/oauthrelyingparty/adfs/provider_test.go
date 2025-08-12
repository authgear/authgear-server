package adfs

import (
	"context"
	"net/http"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"github.com/h2non/gock"

	"github.com/authgear/oauthrelyingparty/pkg/api/oauthrelyingparty"
)

func TestADFS(t *testing.T) {
	Convey("ADFS", t, func() {
		client := &http.Client{}
		gock.InterceptClient(client)
		defer gock.Off()

		deps := oauthrelyingparty.Dependencies{
			ProviderConfig: oauthrelyingparty.ProviderConfig{
				"client_id":                   "client_id",
				"type":                        Type,
				"discovery_document_endpoint": "https://localhost/.well-known/openid-configuration",
			},
			HTTPClient: client,
		}

		g := ADFS{}

		gock.New("https://localhost/.well-known/openid-configuration").
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
		So(u, ShouldEqual, "https://localhost/authorize?client_id=client_id&nonce=nonce&prompt=login&redirect_uri=https%3A%2F%2Flocalhost%2F&response_mode=form_post&response_type=code&scope=openid+profile+email&state=state")
	})
}
