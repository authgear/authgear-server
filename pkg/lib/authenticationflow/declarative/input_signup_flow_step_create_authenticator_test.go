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

		test((&InputSchemaSignupFlowStepCreateAuthenticator{
			OneOf: []*config.AuthenticationFlowSignupFlowOneOf{
				{
					Authentication: config.AuthenticationFlowAuthenticationPrimaryPassword,
				},
				{
					Authentication: config.AuthenticationFlowAuthenticationPrimaryOOBOTPEmail,
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
