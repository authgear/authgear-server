package declarative

import (
	"encoding/json"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/lib/config"
)

func TestInputSchemaStepAccountRecoveryIdentify(t *testing.T) {
	Convey("InputSchemaStepAccountRecoveryIdentify", t, func() {
		test := func(s *InputSchemaStepAccountRecoveryIdentify, expected string) {
			b := s.SchemaBuilder()
			bytes, err := json.Marshal(b)
			So(err, ShouldBeNil)
			So(string(bytes), ShouldEqualJSON, expected)
		}
		var varTrue = true
		dummyBotProtection := &BotProtectionData{
			Enabled: &varTrue,
			Provider: &BotProtectionDataProvider{
				Type: config.BotProtectionProviderTypeCloudflare,
			},
		}
		test(&InputSchemaStepAccountRecoveryIdentify{
			Options: []AccountRecoveryIdentificationOption{
				{
					Identification: config.AuthenticationFlowAccountRecoveryIdentificationEmail,
					BotProtection:  dummyBotProtection,
				},
				{
					Identification: config.AuthenticationFlowAccountRecoveryIdentificationPhone,
					BotProtection:  dummyBotProtection,
				},
			},
		}, `
{
    "oneOf": [
        {
            "properties": {
                "bot_protection": {
                    "allOf": [
                        {
                            "if": {
                                "properties": {
                                    "type": {
                                        "enum": [
                                            "cloudflare",
                                            "recaptchav2"
                                        ]
                                    }
                                },
                                "required": [
                                    "type"
                                ]
                            },
                            "then": {
                                "required": [
                                    "response",
                                    "type"
                                ]
                            }
                        }
                    ],
                    "properties": {
                        "response": {
                            "type": "string"
                        },
                        "type": {
                            "enum": [
                                "cloudflare",
                                "recaptchav2"
                            ],
                            "type": "string"
                        }
                    },
                    "required": [
                        "type"
                    ],
                    "type": "object"
                },
                "identification": {
                    "const": "email"
                },
                "login_id": {
                    "type": "string"
                }
            },
            "required": [
                "identification",
                "login_id",
                "bot_protection"
            ]
        },
        {
            "properties": {
                "bot_protection": {
                    "allOf": [
                        {
                            "if": {
                                "properties": {
                                    "type": {
                                        "enum": [
                                            "cloudflare",
                                            "recaptchav2"
                                        ]
                                    }
                                },
                                "required": [
                                    "type"
                                ]
                            },
                            "then": {
                                "required": [
                                    "response",
                                    "type"
                                ]
                            }
                        }
                    ],
                    "properties": {
                        "response": {
                            "type": "string"
                        },
                        "type": {
                            "enum": [
                                "cloudflare",
                                "recaptchav2"
                            ],
                            "type": "string"
                        }
                    },
                    "required": [
                        "type"
                    ],
                    "type": "object"
                },
                "identification": {
                    "const": "phone"
                },
                "login_id": {
                    "type": "string"
                }
            },
            "required": [
                "identification",
                "login_id",
                "bot_protection"
            ]
        }
    ],
    "type": "object"
}
`)
	})
}
