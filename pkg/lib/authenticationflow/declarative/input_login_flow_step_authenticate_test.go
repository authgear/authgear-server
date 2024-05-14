package declarative

import (
	"encoding/json"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/validation"
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

		Convey("options", func() {
			test((&InputSchemaLoginFlowStepAuthenticate{
				Options: []AuthenticateOption{
					{
						Authentication: config.AuthenticationFlowAuthenticationPrimaryPassword,
					},
					{
						Authentication: config.AuthenticationFlowAuthenticationPrimaryPasskey,
					},
					{
						Authentication: config.AuthenticationFlowAuthenticationSecondaryPassword,
					},
					{
						Authentication: config.AuthenticationFlowAuthenticationSecondaryTOTP,
					},
					{
						Authentication: config.AuthenticationFlowAuthenticationPrimaryOOBOTPEmail,
					},
					{
						Authentication: config.AuthenticationFlowAuthenticationPrimaryOOBOTPSMS,
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
    "properties": {
        "request_device_token": {
            "const": false,
            "type": "boolean"
        }
    },
    "oneOf": [
        {
            "properties": {
                "authentication": {
                    "const": "primary_password"
                },
                "password": {
                    "type": "string"
                }
            },
            "required": [
                "authentication",
                "password"
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
                            "format": "x_base64_url",
                            "type": "string"
                        },
                        "response": {
                            "properties": {
                                "authenticatorData": {
                                    "format": "x_base64_url",
                                    "type": "string"
                                },
                                "clientDataJSON": {
                                    "format": "x_base64_url",
                                    "type": "string"
                                },
                                "signature": {
                                    "format": "x_base64_url",
                                    "type": "string"
                                },
                                "userHandle": {
                                    "format": "x_base64_url",
                                    "type": "string"
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
                "authentication": {
                    "const": "primary_passkey"
                }
            },
            "required": [
                "authentication",
                "assertion_response"
            ]
        },
        {
            "properties": {
                "authentication": {
                    "const": "secondary_password"
                },
                "password": {
                    "type": "string"
                }
            },
            "required": [
                "authentication",
                "password"
            ]
        },
        {
            "properties": {
                "authentication": {
                    "const": "secondary_totp"
                },
                "code": {
                    "type": "string"
                }
            },
            "required": [
                "authentication",
                "code"
            ]
        },
        {
            "properties": {
                "authentication": {
                    "const": "primary_oob_otp_email"
                },
                "index": {
                    "const": 4,
                    "type": "integer"
                }
            },
            "required": [
                "authentication",
                "index"
            ]
        },
        {
            "properties": {
                "authentication": {
                    "const": "primary_oob_otp_sms"
                },
                "index": {
                    "const": 5,
                    "type": "integer"
                }
            },
            "required": [
                "authentication",
                "index"
            ]
        },
        {
            "properties": {
                "authentication": {
                    "const": "secondary_oob_otp_email"
                },
                "index": {
                    "const": 6,
                    "type": "integer"
                }
            },
            "required": [
                "authentication",
                "index"
            ]
        },
        {
            "properties": {
                "authentication": {
                    "const": "secondary_oob_otp_sms"
                },
                "index": {
                    "const": 7,
                    "type": "integer"
                }
            },
            "required": [
                "authentication",
                "index"
            ]
        },
        {
            "properties": {
                "authentication": {
                    "const": "recovery_code"
                },
                "recovery_code": {
                    "type": "string"
                }
            },
            "required": [
                "authentication",
                "recovery_code"
            ]
        }
    ]
}
`)
		})
	})
}
