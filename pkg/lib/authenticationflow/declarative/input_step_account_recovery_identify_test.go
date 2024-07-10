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
		var dummyBotProtectionCfg = &config.BotProtectionConfig{
			Enabled: true,
			Provider: &config.BotProtectionProvider{
				Type: config.BotProtectionProviderTypeCloudflare,
			},
		}
		test(&InputSchemaStepAccountRecoveryIdentify{
			BotProtectionCfg: dummyBotProtectionCfg,
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
                    "properties": {
                        "response": {
                            "type": "string"
                        },
                        "type": {
                            "const": "cloudflare"
                        }
                    },
                    "required": [
                        "type",
                        "response"
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
                "bot_protection",
                "login_id"
            ]
        },
        {
            "properties": {
                "bot_protection": {
                    "properties": {
                        "response": {
                            "type": "string"
                        },
                        "type": {
                            "const": "cloudflare"
                        }
                    },
                    "required": [
                        "type",
                        "response"
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
                "bot_protection",
                "login_id"
            ]
        }
    ],
    "type": "object"
}
`)
	})
}
