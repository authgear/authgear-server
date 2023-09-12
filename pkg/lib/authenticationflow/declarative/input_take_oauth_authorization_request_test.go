package declarative

import (
	"encoding/json"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/lib/config"
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
			OAuthCandidates: []IdentificationCandidate{
				{
					Identification: config.AuthenticationFlowIdentificationOAuth,
					Alias:          "google",
				},
				{
					Identification: config.AuthenticationFlowIdentificationOAuth,
					Alias:          "wechat_mobile",
					WechatAppType:  config.OAuthSSOWeChatAppTypeMobile,
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
        "state": {
            "type": "string"
        }
    },
    "required": [
        "alias",
        "state",
        "redirect_uri"
    ],
    "type": "object"
}
		`)
	})
}
