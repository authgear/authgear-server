package declarative

import (
	"encoding/json"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/lib/config"
)

func TestInputSchemaStepIdentify(t *testing.T) {
	Convey("InputSchemaStepIdentify", t, func() {
		test := func(s *InputSchemaStepIdentify, expected string) {
			b := s.SchemaBuilder()
			bytes, err := json.Marshal(b)
			So(err, ShouldBeNil)
			So(string(bytes), ShouldEqualJSON, expected)
		}

		test(&InputSchemaStepIdentify{
			Identifications: []config.AuthenticationFlowIdentification{
				config.AuthenticationFlowIdentificationEmail,
				config.AuthenticationFlowIdentificationPhone,
				config.AuthenticationFlowIdentificationUsername,
				config.AuthenticationFlowIdentificationOAuth,
			},
			OAuthConfig: &config.OAuthSSOConfig{
				Providers: []config.OAuthSSOProviderConfig{
					{
						Alias: "google",
					},
				},
			},
		}, `
{
    "oneOf": [
        {
            "properties": {
                "identification": {
                    "const": "email"
                },
                "login_id": {
                    "type": "string"
                }
            },
            "required": [
                "identification",
                "login_id"
            ]
        },
        {
            "properties": {
                "identification": {
                    "const": "phone"
                },
                "login_id": {
                    "type": "string"
                }
            },
            "required": [
                "identification",
                "login_id"
            ]
        },
        {
            "properties": {
                "identification": {
                    "const": "username"
                },
                "login_id": {
                    "type": "string"
                }
            },
            "required": [
                "identification",
                "login_id"
            ]
        },
        {
            "properties": {
                "alias": {
                    "enum": [
                        "google"
                    ],
                    "type": "string"
                },
                "identification": {
                    "const": "oauth"
                },
                "redirect_uri": {
                    "type": "string"
                }
            },
            "required": [
                "identification",
                "redirect_uri",
                "alias"
            ]
        }
    ],
    "type": "object"
}
		`)
	})
}
