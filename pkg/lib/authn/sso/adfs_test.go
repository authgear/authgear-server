package sso

import (
	"net/http"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"gopkg.in/h2non/gock.v1"

	"github.com/authgear/authgear-server/pkg/api/oauthrelyingparty"
	"github.com/authgear/authgear-server/pkg/lib/oauthrelyingparty/adfs"
)

func TestADFSImpl(t *testing.T) {
	Convey("ADFSImpl", t, func() {
		client := OAuthHTTPClient{&http.Client{}}
		gock.InterceptClient(client.Client)
		defer gock.Off()

		g := &ADFSImpl{
			ProviderConfig: oauthrelyingparty.ProviderConfig{
				"client_id":                   "client_id",
				"type":                        adfs.Type,
				"discovery_document_endpoint": "https://localhost/.well-known/openid-configuration",
			},
			HTTPClient: client,
		}

		gock.New("https://localhost/.well-known/openid-configuration").
			Reply(200).
			BodyString(`
{
  "authorization_endpoint": "https://localhost/authorize"
}
			`)
		defer func() { gock.Flush() }()

		u, err := g.GetAuthorizationURL(GetAuthorizationURLOptions{
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
