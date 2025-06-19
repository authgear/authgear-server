package declarative

import (
	"encoding/json"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/oauthrelyingparty/wechat"
)

func TestInputSchemaTakeOAuthAuthorizationRequest(t *testing.T) {
	Convey("InputSchemaTakeOAuthAuthorizationRequest", t, func() {
		test := func(s *InputSchemaTakeOAuthAuthorizationRequest, expected string) {
			b := s.SchemaBuilder()
			bytes, err := json.Marshal(b)
			So(err, ShouldBeNil)
			So(string(bytes), ShouldEqualJSON, expected)
		}

		test(&InputSchemaTakeOAuthAuthorizationRequest{
			OAuthOptions: []IdentificationOption{
				{
					Identification: model.AuthenticationFlowIdentificationOAuth,
					Alias:          "google",
				},
				{
					Identification: model.AuthenticationFlowIdentificationOAuth,
					Alias:          "wechat_mobile",
					WechatAppType:  wechat.AppTypeMobile,
				},
			},
		}, `
{
    "properties": {
        "alias": {
            "enum": [
                "google",
                "wechat_mobile"
            ],
            "type": "string"
        },
        "redirect_uri": {
            "type": "string",
            "format": "uri"
        },
        "response_mode": {
            "type": "string",
            "enum": [
                "form_post",
                "query"
            ]
        }
    },
    "required": [
        "alias",
        "redirect_uri"
    ],
    "type": "object"
}
		`)
	})
}
