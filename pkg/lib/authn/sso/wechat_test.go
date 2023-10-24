package sso

import (
	"testing"

	"github.com/authgear/authgear-server/pkg/lib/config"
	. "github.com/smartystreets/goconvey/convey"
)

func TestWechatImpl(t *testing.T) {
	Convey("WechatImpl", t, func() {

		g := &WechatImpl{
			ProviderConfig: config.OAuthSSOProviderConfig{
				ClientID: "client_id",
				Type:     config.OAuthSSOProviderTypeWechat,
			},
		}

		u, err := g.GetAuthURL(GetAuthURLParam{
			Nonce:  "nonce",
			State:  "state",
			Prompt: []string{"login"},
		})
		So(err, ShouldBeNil)
		So(u, ShouldEqual, "https://open.weixin.qq.com/connect/oauth2/authorize?appid=client_id&redirect_uri=&response_type=code&scope=snsapi_userinfo&state=state")
	})
}
