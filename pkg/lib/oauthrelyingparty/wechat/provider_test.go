package wechat

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/api/oauthrelyingparty"
)

func TestWechat(t *testing.T) {
	Convey("Wechat", t, func() {
		deps := oauthrelyingparty.Dependencies{
			ProviderConfig: oauthrelyingparty.ProviderConfig{
				"client_id": "client_id",
				"type":      Type,
			},
		}

		g := Wechat{}

		u, err := g.GetAuthorizationURL(deps, oauthrelyingparty.GetAuthorizationURLOptions{
			Nonce:  "nonce",
			State:  "state",
			Prompt: []string{"login"},
		})
		So(err, ShouldBeNil)
		So(u, ShouldEqual, "https://open.weixin.qq.com/connect/oauth2/authorize?appid=client_id&redirect_uri=&response_type=code&scope=snsapi_userinfo&state=state")
	})
}
