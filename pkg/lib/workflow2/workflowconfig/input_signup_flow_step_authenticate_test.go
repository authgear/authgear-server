package workflowconfig

import (
	"encoding/json"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

func TestInputSchemaSignupFlowStepAuthenticate(t *testing.T) {
	Convey("InputSchemaSignupFlowStepAuthenticate", t, func() {
		test := func(b validation.SchemaBuilder, expected string) {
			bytes, err := json.Marshal(b)
			So(err, ShouldBeNil)
			So(string(bytes), ShouldEqualJSON, expected)
		}

		test((&InputSchemaSignupFlowStepAuthenticate{
			OneOf: []*config.WorkflowSignupFlowOneOf{
				{
					Authentication: config.WorkflowAuthenticationMethodPrimaryPassword,
				},
				{
					Authentication: config.WorkflowAuthenticationMethodPrimaryPasskey,
				},
				{
					Authentication: config.WorkflowAuthenticationMethodPrimaryOOBOTPEmail,
				},
				{
					Authentication: config.WorkflowAuthenticationMethodPrimaryOOBOTPSMS,
					TargetStep:     "step",
				},
				{
					Authentication: config.WorkflowAuthenticationMethodSecondaryPassword,
				},
				{
					Authentication: config.WorkflowAuthenticationMethodSecondaryTOTP,
				},
				{
					Authentication: config.WorkflowAuthenticationMethodSecondaryOOBOTPEmail,
				},
				{
					Authentication: config.WorkflowAuthenticationMethodSecondaryOOBOTPSMS,
				},
				{
					Authentication: config.WorkflowAuthenticationMethodRecoveryCode,
				},
				{
					Authentication: config.WorkflowAuthenticationMethodDeviceToken,
				},
			},
		}).SchemaBuilder(), `
{
    "type": "object",
    "oneOf": [
        {
            "properties": {
                "authentication_method": {
                    "const": "primary_password"
                },
                "new_password": {
                    "type": "string"
                }
            },
            "required": [
                "authentication_method",
                "new_password"
            ]
        },
        {
            "properties": {
                "authentication_method": {
                    "const": "primary_oob_otp_email"
                },
                "target": {
                    "type": "string"
                }
            },
            "required": [
                "authentication_method",
                "target"
            ]
        },
        {
            "properties": {
                "authentication_method": {
                    "const": "primary_oob_otp_sms"
                }
            },
            "required": [
                "authentication_method"
            ]
        },
        {
            "properties": {
                "authentication_method": {
                    "const": "secondary_password"
                },
                "new_password": {
                    "type": "string"
                }
            },
            "required": [
                "authentication_method",
                "new_password"
            ]
        },
        {
            "properties": {
                "authentication_method": {
                    "const": "secondary_totp"
                }
            },
            "required": [
                "authentication_method"
            ]
        },
        {
            "properties": {
                "authentication_method": {
                    "const": "secondary_oob_otp_email"
                },
                "target": {
                    "type": "string"
                }
            },
            "required": [
                "authentication_method",
                "target"
            ]
        },
        {
            "properties": {
                "authentication_method": {
                    "const": "secondary_oob_otp_sms"
                },
                "target": {
                    "type": "string"
                }
            },
            "required": [
                "authentication_method",
                "target"
            ]
        }
    ]
}
		`)
	})
}
