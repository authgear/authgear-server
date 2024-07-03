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
		dummyBotProtectionCfg := &config.AuthenticationFlowBotProtection{
			Mode: config.AuthenticationFlowBotProtectionModeAlways,
			Provider: &config.AuthenticationFlowBotProtectionProvider{
				Type: config.BotProtectionProviderTypeCloudflare,
			},
		}

		test((&InputSchemaSignupFlowStepCreateAuthenticator{
			OneOf: []*config.AuthenticationFlowSignupFlowOneOf{
				{
					Authentication: config.AuthenticationFlowAuthenticationPrimaryPassword,
					BotProtection:  dummyBotProtectionCfg,
				},
				{
					Authentication: config.AuthenticationFlowAuthenticationPrimaryOOBOTPEmail,
					BotProtection:  dummyBotProtectionCfg,
				},
				{
					Authentication: config.AuthenticationFlowAuthenticationPrimaryOOBOTPSMS,
					TargetStep:     "step",
					BotProtection:  dummyBotProtectionCfg,
				},
				{
					Authentication: config.AuthenticationFlowAuthenticationSecondaryPassword,
					BotProtection:  dummyBotProtectionCfg,
				},
				{
					Authentication: config.AuthenticationFlowAuthenticationSecondaryTOTP,
					BotProtection:  dummyBotProtectionCfg,
				},
				{
					Authentication: config.AuthenticationFlowAuthenticationSecondaryOOBOTPEmail,
					BotProtection:  dummyBotProtectionCfg,
				},
				{
					Authentication: config.AuthenticationFlowAuthenticationSecondaryOOBOTPSMS,
					BotProtection:  dummyBotProtectionCfg,
				},
				{
					Authentication: config.AuthenticationFlowAuthenticationRecoveryCode,
					BotProtection:  dummyBotProtectionCfg,
				},
				{
					Authentication: config.AuthenticationFlowAuthenticationDeviceToken,
					BotProtection:  dummyBotProtectionCfg,
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
                "new_password": {
                    "type": "string"
                }
            },
            "required": [
                "authentication",
                "bot_protection",
                "new_password"
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
                "target": {
                    "type": "string"
                }
            },
            "required": [
                "authentication",
                "bot_protection",
                "target"
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
                }
            },
            "required": [
                "authentication",
                "bot_protection"
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
                "new_password": {
                    "type": "string"
                }
            },
            "required": [
                "authentication",
                "bot_protection",
                "new_password"
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
                }
            },
            "required": [
                "authentication",
                "bot_protection"
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
                "target": {
                    "type": "string"
                }
            },
            "required": [
                "authentication",
                "bot_protection",
                "target"
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
                "target": {
                    "type": "string"
                }
            },
            "required": [
                "authentication",
                "bot_protection",
                "target"
            ]
        }
    ]
}
		`)
	})
}
