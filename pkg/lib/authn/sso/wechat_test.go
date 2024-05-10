package sso

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/api/oauthrelyingparty"
	"github.com/authgear/authgear-server/pkg/lib/oauthrelyingparty/wechat"
)

func TestWechatImpl(t *testing.T) {
	Convey("WechatImpl", t, func() {

		g := &WechatImpl{
			ProviderConfig: oauthrelyingparty.ProviderConfig{
				"client_id": "client_id",
				"type":      wechat.Type,
			},
			HTTPClient: OAuthHTTPClient{},
		}

		u, err := g.GetAuthorizationURL(oauthrelyingparty.GetAuthorizationURLOptions{
			Nonce:  "nonce",
			State:  "state",
			Prompt: []string{"login"},
		})
		So(err, ShouldBeNil)
		So(u, ShouldEqual, "https://open.weixin.qq.com/connect/oauth2/authorize?appid=client_id&redirect_uri=&response_type=code&scope=snsapi_userinfo&state=state")
	})
}
