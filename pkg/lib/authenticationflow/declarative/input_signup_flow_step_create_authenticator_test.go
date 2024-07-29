package declarative

import (
	"encoding/json"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

func TestInputSchemaSignupFlowStepCreateAuthenticator(t *testing.T) {
	Convey("InputSchemaSignupFlowStepCreateAuthenticator", t, func() {
		test := func(b validation.SchemaBuilder, expected string) {
			bytes, err := json.Marshal(b)
			So(err, ShouldBeNil)
			So(string(bytes), ShouldEqualJSON, expected)
		}

		var dummyBotProtectionCfg = &config.BotProtectionConfig{
			Enabled: true,
			Provider: &config.BotProtectionProvider{
				Type: config.BotProtectionProviderTypeCloudflare,
			},
		}

		var dummyBotProtectionData = &config.AuthenticationFlowBotProtection{
			Mode: config.AuthenticationFlowBotProtectionModeAlways,
			Provider: &config.AuthenticationFlowBotProtectionProvider{
				Type: config.BotProtectionProviderTypeCloudflare,
			},
		}

		test((&InputSchemaSignupFlowStepCreateAuthenticator{
			ShouldBypassBotProtection: false,
			BotProtectionCfg:          dummyBotProtectionCfg,
			OneOf: []*config.AuthenticationFlowSignupFlowOneOf{
				{
					Authentication: config.AuthenticationFlowAuthenticationPrimaryPassword,
				},
				{
					Authentication: config.AuthenticationFlowAuthenticationPrimaryOOBOTPEmail,
					BotProtection:  dummyBotProtectionData,
				},
				{
					Authentication: config.AuthenticationFlowAuthenticationPrimaryOOBOTPSMS,
					TargetStep:     "step",
				},
				{
					Authentication: config.AuthenticationFlowAuthenticationSecondaryPassword,
				},
				{
					Authentication: config.AuthenticationFlowAuthenticationSecondaryTOTP,
				},
				{
					Authentication: config.AuthenticationFlowAuthenticationSecondaryOOBOTPEmail,
				},
				{
					Authentication: config.AuthenticationFlowAuthenticationSecondaryOOBOTPSMS,
				},
				{
					Authentication: config.AuthenticationFlowAuthenticationRecoveryCode,
				},
				{
					Authentication: config.AuthenticationFlowAuthenticationDeviceToken,
				},
			},
		}).SchemaBuilder(), `
{
    "type": "object",
    "oneOf": [
        {
            "properties": {
                "authentication": {
                    "const": "primary_password"
                },
                "new_password": {
                    "type": "string"
                }
            },
            "required": [
                "authentication",
                "new_password"
            ]
        },
        {
            "properties": {
                "authentication": {
                    "const": "primary_oob_otp_email"
                },
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
                "target": {
                    "type": "string"
                }
            },
            "required": [
                "authentication",
                "target",
                "bot_protection"
            ]
        },
        {
            "properties": {
                "authentication": {
                    "const": "primary_oob_otp_sms"
                }
            },
            "required": [
                "authentication"
            ]
        },
        {
            "properties": {
                "authentication": {
                    "const": "secondary_password"
                },
                "new_password": {
                    "type": "string"
                }
            },
            "required": [
                "authentication",
                "new_password"
            ]
        },
        {
            "properties": {
                "authentication": {
                    "const": "secondary_totp"
                }
            },
            "required": [
                "authentication"
            ]
        },
        {
            "properties": {
                "authentication": {
                    "const": "secondary_oob_otp_email"
                },
                "target": {
                    "type": "string"
                }
            },
            "required": [
                "authentication",
                "target"
            ]
        },
        {
            "properties": {
                "authentication": {
                    "const": "secondary_oob_otp_sms"
                },
                "target": {
                    "type": "string"
                }
            },
            "required": [
                "authentication",
                "target"
            ]
        }
    ]
}
		`)
	})
}
