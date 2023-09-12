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
			OAuthConfig: &config.OAuthSSOConfig{
				Providers: []config.OAuthSSOProviderConfig{
					{
						Alias: "google",
					},
				},
			},
		}, `
{
    "properties": {
        "alias": {
            "enum": [
                "google"
            ],
            "type": "string"
        },
        "redirect_uri": {
            "type": "string"
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
