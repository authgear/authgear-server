package workflowconfig

import (
	"encoding/json"
	"testing"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/validation"
	. "github.com/smartystreets/goconvey/convey"
)

func TestInputSchemaLoginFlowStepAuthenticate(t *testing.T) {
	Convey("InputSchemaLoginFlowStepAuthenticate", t, func() {
		test := func(b validation.SchemaBuilder, expected string) {
			bytes, err := json.Marshal(b)
			So(err, ShouldBeNil)
			So(string(bytes), ShouldEqualJSON, expected)
		}

		Convey("device token is false when it is not enabled", func() {
			test((&InputSchemaLoginFlowStepAuthenticate{}).SchemaBuilder(), `
{
    "type": "object",
    "properties": {
        "request_device_token": {
            "const": false,
            "type": "boolean"
        }
    }
}
`)
		})

		Convey("device token is type boolean when it is enabled", func() {
			test((&InputSchemaLoginFlowStepAuthenticate{
				DeviceTokenEnabled: true,
			}).SchemaBuilder(), `
{
    "type": "object",
    "properties": {
        "request_device_token": {
            "type": "boolean"
        }
    }
}
`)
		})

		Convey("candidates", func() {
			test((&InputSchemaLoginFlowStepAuthenticate{
				Candidates: []UseAuthenticationCandidate{
					{
						AuthenticationMethod: config.WorkflowAuthenticationMethodPrimaryPassword,
					},
					{
						AuthenticationMethod: config.WorkflowAuthenticationMethodSecondaryPassword,
					},
					{
						AuthenticationMethod: config.WorkflowAuthenticationMethodSecondaryTOTP,
					},
					{
						AuthenticationMethod: config.WorkflowAuthenticationMethodPrimaryOOBOTPEmail,
					},
					{
						AuthenticationMethod: config.WorkflowAuthenticationMethodPrimaryOOBOTPSMS,
					},
					{
						AuthenticationMethod: config.WorkflowAuthenticationMethodSecondaryOOBOTPEmail,
					},
					{
						AuthenticationMethod: config.WorkflowAuthenticationMethodSecondaryOOBOTPSMS,
					},
					{
						AuthenticationMethod: config.WorkflowAuthenticationMethodRecoveryCode,
					},
					{
						AuthenticationMethod: config.WorkflowAuthenticationMethodDeviceToken,
					},
				},
			}).SchemaBuilder(), `
{
    "type": "object",
    "properties": {
        "request_device_token": {
            "const": false,
            "type": "boolean"
        }
    },
    "oneOf": [
        {
            "properties": {
                "authentication_method": {
                    "const": "primary_password"
                },
                "password": {
                    "type": "string"
                }
            },
            "required": [
                "authentication_method",
                "password"
            ]
        },
        {
            "properties": {
                "authentication_method": {
                    "const": "secondary_password"
                },
                "password": {
                    "type": "string"
                }
            },
            "required": [
                "authentication_method",
                "password"
            ]
        },
        {
            "properties": {
                "authentication_method": {
                    "const": "secondary_totp"
                },
                "code": {
                    "type": "string"
                }
            },
            "required": [
                "authentication_method",
                "code"
            ]
        },
        {
            "properties": {
                "authentication_method": {
                    "const": "primary_oob_otp_email"
                },
                "index": {
                    "const": 3,
                    "type": "integer"
                }
            },
            "required": [
                "authentication_method",
                "index"
            ]
        },
        {
            "properties": {
                "authentication_method": {
                    "const": "primary_oob_otp_sms"
                },
                "index": {
                    "const": 4,
                    "type": "integer"
                }
            },
            "required": [
                "authentication_method",
                "index"
            ]
        },
        {
            "properties": {
                "authentication_method": {
                    "const": "secondary_oob_otp_email"
                },
                "index": {
                    "const": 5,
                    "type": "integer"
                }
            },
            "required": [
                "authentication_method",
                "index"
            ]
        },
        {
            "properties": {
                "authentication_method": {
                    "const": "secondary_oob_otp_sms"
                },
                "index": {
                    "const": 6,
                    "type": "integer"
                }
            },
            "required": [
                "authentication_method",
                "index"
            ]
        },
        {
            "properties": {
                "authentication_method": {
                    "const": "recovery_code"
                },
                "recovery_code": {
                    "type": "string"
                }
            },
            "required": [
                "authentication_method",
                "recovery_code"
            ]
        }
    ]
}
`)
		})
	})
}
