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
		Convey("options with bot protections should require bot protection input", func() {
			var varTrue = true
			var dummyBotProtectionData = &BotProtectionData{
				Enabled: &varTrue,
				Provider: &BotProtectionDataProvider{
					Type: config.BotProtectionProviderTypeCloudflare,
				},
			}
			test((&InputSchemaLoginFlowStepAuthenticate{
				Options: []AuthenticateOption{
					{
						Authentication: config.AuthenticationFlowAuthenticationPrimaryPassword,
						BotProtection:  dummyBotProtectionData,
					},
					{
						Authentication: config.AuthenticationFlowAuthenticationPrimaryPasskey,
						BotProtection:  dummyBotProtectionData,
					},
					{
						Authentication: config.AuthenticationFlowAuthenticationSecondaryPassword,
						BotProtection:  dummyBotProtectionData,
					},
					{
						Authentication: config.AuthenticationFlowAuthenticationSecondaryTOTP,
						BotProtection:  dummyBotProtectionData,
					},
					{
						Authentication: config.AuthenticationFlowAuthenticationPrimaryOOBOTPEmail,
						BotProtection:  dummyBotProtectionData,
					},
					{
						Authentication: config.AuthenticationFlowAuthenticationPrimaryOOBOTPSMS,
						BotProtection:  dummyBotProtectionData,
					},
					{
						Authentication: config.AuthenticationFlowAuthenticationSecondaryOOBOTPEmail,
						BotProtection:  dummyBotProtectionData,
					},
					{
						Authentication: config.AuthenticationFlowAuthenticationSecondaryOOBOTPSMS,
						BotProtection:  dummyBotProtectionData,
					},
					{
						Authentication: config.AuthenticationFlowAuthenticationRecoveryCode,
						BotProtection:  dummyBotProtectionData,
					},
					{
						Authentication: config.AuthenticationFlowAuthenticationDeviceToken,
						BotProtection:  dummyBotProtectionData,
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
                "password": {
                    "type": "string"
                }
            },
            "required": [
                "authentication",
                "bot_protection",
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
                },
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
                }
            },
            "required": [
                "authentication",
                "bot_protection",
                "assertion_response"
            ]
        },
        {
            "properties": {
                "authentication": {
                    "const": "secondary_password"
                },
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
                "password": {
                    "type": "string"
                }
            },
            "required": [
                "authentication",
                "bot_protection",
                "password"
            ]
        },
        {
            "properties": {
                "authentication": {
                    "const": "secondary_totp"
                },
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
                "code": {
                    "type": "string"
                }
            },
            "required": [
                "authentication",
                "bot_protection",
                "code"
            ]
        },
        {
            "properties": {
                "authentication": {
                    "const": "primary_oob_otp_email"
                },
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
                "index": {
                    "const": 4,
                    "type": "integer"
                }
            },
            "required": [
                "authentication",
                "bot_protection",
                "index"
            ]
        },
        {
            "properties": {
                "authentication": {
                    "const": "primary_oob_otp_sms"
                },
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
                "index": {
                    "const": 5,
                    "type": "integer"
                }
            },
            "required": [
                "authentication",
                "bot_protection",
                "index"
            ]
        },
        {
            "properties": {
                "authentication": {
                    "const": "secondary_oob_otp_email"
                },
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
                "index": {
                    "const": 6,
                    "type": "integer"
                }
            },
            "required": [
                "authentication",
                "bot_protection",
                "index"
            ]
        },
        {
            "properties": {
                "authentication": {
                    "const": "secondary_oob_otp_sms"
                },
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
                "index": {
                    "const": 7,
                    "type": "integer"
                }
            },
            "required": [
                "authentication",
                "bot_protection",
                "index"
            ]
        },
        {
            "properties": {
                "authentication": {
                    "const": "recovery_code"
                },
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
                "recovery_code": {
                    "type": "string"
                }
            },
            "required": [
                "authentication",
                "bot_protection",
                "recovery_code"
            ]
        }
    ]
}
`)
		})
		Convey("only options w/ bot protections should require bot protection input, options w/o bot protections should not require bot protection input", func() {
			var varTrue = true
			var dummyBotProtectionData = &BotProtectionData{
				Enabled: &varTrue,
				Provider: &BotProtectionDataProvider{
					Type: config.BotProtectionProviderTypeCloudflare,
				},
			}
			test((&InputSchemaLoginFlowStepAuthenticate{
				Options: []AuthenticateOption{
					{
						Authentication: config.AuthenticationFlowAuthenticationPrimaryPassword,
					},
					{
						Authentication: config.AuthenticationFlowAuthenticationPrimaryPasskey,
						BotProtection:  dummyBotProtectionData,
					},
					{
						Authentication: config.AuthenticationFlowAuthenticationSecondaryPassword,
						BotProtection:  dummyBotProtectionData,
					},
					{
						Authentication: config.AuthenticationFlowAuthenticationSecondaryTOTP,
						BotProtection:  dummyBotProtectionData,
					},
					{
						Authentication: config.AuthenticationFlowAuthenticationPrimaryOOBOTPEmail,
						BotProtection:  dummyBotProtectionData,
					},
					{
						Authentication: config.AuthenticationFlowAuthenticationPrimaryOOBOTPSMS,
						BotProtection:  dummyBotProtectionData,
					},
					{
						Authentication: config.AuthenticationFlowAuthenticationSecondaryOOBOTPEmail,
						BotProtection:  dummyBotProtectionData,
					},
					{
						Authentication: config.AuthenticationFlowAuthenticationSecondaryOOBOTPSMS,
						BotProtection:  dummyBotProtectionData,
					},
					{
						Authentication: config.AuthenticationFlowAuthenticationRecoveryCode,
					},
					{
						Authentication: config.AuthenticationFlowAuthenticationDeviceToken,
						BotProtection:  dummyBotProtectionData,
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
                },
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
                }
            },
            "required": [
                "authentication",
                "bot_protection",
                "assertion_response"
            ]
        },
        {
            "properties": {
                "authentication": {
                    "const": "secondary_password"
                },
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
                "password": {
                    "type": "string"
                }
            },
            "required": [
                "authentication",
                "bot_protection",
                "password"
            ]
        },
        {
            "properties": {
                "authentication": {
                    "const": "secondary_totp"
                },
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
                "code": {
                    "type": "string"
                }
            },
            "required": [
                "authentication",
                "bot_protection",
                "code"
            ]
        },
        {
            "properties": {
                "authentication": {
                    "const": "primary_oob_otp_email"
                },
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
                "index": {
                    "const": 4,
                    "type": "integer"
                }
            },
            "required": [
                "authentication",
                "bot_protection",
                "index"
            ]
        },
        {
            "properties": {
                "authentication": {
                    "const": "primary_oob_otp_sms"
                },
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
                "index": {
                    "const": 5,
                    "type": "integer"
                }
            },
            "required": [
                "authentication",
                "bot_protection",
                "index"
            ]
        },
        {
            "properties": {
                "authentication": {
                    "const": "secondary_oob_otp_email"
                },
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
                "index": {
                    "const": 6,
                    "type": "integer"
                }
            },
            "required": [
                "authentication",
                "bot_protection",
                "index"
            ]
        },
        {
            "properties": {
                "authentication": {
                    "const": "secondary_oob_otp_sms"
                },
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
                "index": {
                    "const": 7,
                    "type": "integer"
                }
            },
            "required": [
                "authentication",
                "bot_protection",
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
