package declarative

import (
	"encoding/json"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/oauthrelyingparty/wechat"
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
			Options: []IdentificationOption{
				{
					Identification: config.AuthenticationFlowIdentificationEmail,
				},
				{
					Identification: config.AuthenticationFlowIdentificationPhone,
				},
				{
					Identification: config.AuthenticationFlowIdentificationUsername,
				},
				{
					Identification: config.AuthenticationFlowIdentificationOAuth,
					Alias:          "google",
				},
				{
					Identification: config.AuthenticationFlowIdentificationOAuth,
					Alias:          "wechat_mobile",
					WechatAppType:  wechat.AppTypeMobile,
				},
				{
					Identification: config.AuthenticationFlowIdentificationPasskey,
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
                    "const": "google",
                    "type": "string"
                },
                "identification": {
                    "const": "oauth"
                },
                "redirect_uri": {
                    "format": "uri",
                    "type": "string"
                },
                "response_mode": {
                    "type": "string",
                    "enum": ["form_post", "query"]
                }
            },
            "required": [
                "identification",
                "redirect_uri",
                "alias"
            ]
        },
        {
            "properties": {
                "alias": {
                    "const": "wechat_mobile",
                    "type": "string"
                },
                "identification": {
                    "const": "oauth"
                },
                "redirect_uri": {
                    "format": "uri",
                    "type": "string"
                },
                "response_mode": {
                    "type": "string",
                    "enum": ["form_post", "query"]
                }
            },
            "required": [
                "identification",
                "redirect_uri",
                "alias"
            ]
        },
        {
            "properties": {
                "assertion_response": {
                    "properties": {
                        "clientExtensionResults": {
                            "type": "object"
                        },
                        "id": {
                            "type": "string"
                        },
                        "rawId": {
                            "type": "string",
                            "format": "x_base64_url"
                        },
                        "response": {
                            "properties": {
                                "authenticatorData": {
                                    "type": "string",
                                    "format": "x_base64_url"
                                },
                                "clientDataJSON": {
                                    "type": "string",
                                    "format": "x_base64_url"
                                },
                                "signature": {
                                    "type": "string",
                                    "format": "x_base64_url"
                                },
                                "userHandle": {
                                    "type": "string",
                                    "format": "x_base64_url"
                                }
                            },
                            "required": [
                                "clientDataJSON",
                                "authenticatorData",
                                "signature"
                            ],
                            "type": "object"
                        },
                        "type": {
                            "type": "string"
                        }
                    },
                    "required": [
                        "id",
                        "type",
                        "rawId",
                        "response"
                    ],
                    "type": "object"
                },
                "identification": {
                    "const": "passkey"
                }
            },
            "required": [
                "identification",
                "assertion_response"
            ]
        }
    ],
    "type": "object"
}
`)
	})
}
