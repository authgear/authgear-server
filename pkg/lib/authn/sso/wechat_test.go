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
			URLProvider: mockWechatURLProvider{},
		}

		u, err := g.GetAuthURL(GetAuthURLParam{
			Nonce:  "nonce",
			State:  "state",
			Prompt: []string{"login"},
		})
		So(err, ShouldBeNil)
		So(u, ShouldEqual, "https://localhost/wechat/authorize?x_auth_url=https%3A%2F%2Fopen.weixin.qq.com%2Fconnect%2Foauth2%2Fauthorize%3Fappid%3Dclient_id%26redirect_uri%3Dhttps%253A%252F%252Flocalhost%252Fwechat%252Fcallback%26response_type%3Dcode%26scope%3Dsnsapi_userinfo%26state%3Dstate")
	})
}
