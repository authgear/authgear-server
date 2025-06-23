package wechat

import (
	"context"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/oauthrelyingparty/pkg/api/oauthrelyingparty"
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

		ctx := context.Background()
		u, err := g.GetAuthorizationURL(ctx, deps, oauthrelyingparty.GetAuthorizationURLOptions{
			Nonce:  "nonce",
			State:  "state",
			Prompt: []string{"login"},
		})
		So(err, ShouldBeNil)
		So(u, ShouldEqual, "https://open.weixin.qq.com/connect/qrconnect?appid=client_id&redirect_uri=&response_type=code&scope=snsapi_login&state=state#wechat_redirect")
	})
}
