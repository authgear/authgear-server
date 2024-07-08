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
		test(&InputSchemaStepIdentify{
			BotProtectionCfg: dummyBotProtectionCfg,
			Options: []IdentificationOption{
				{
					Identification: config.AuthenticationFlowIdentificationEmail,
					BotProtection:  dummyBotProtection,
				},
				{
					Identification: config.AuthenticationFlowIdentificationPhone,
					BotProtection:  dummyBotProtection,
				},
				{
					Identification: config.AuthenticationFlowIdentificationUsername,
					BotProtection:  dummyBotProtection,
				},
				{
					Identification: config.AuthenticationFlowIdentificationOAuth,
					Alias:          "google",
					BotProtection:  dummyBotProtection,
				},
				{
					Identification: config.AuthenticationFlowIdentificationOAuth,
					Alias:          "wechat_mobile",
					WechatAppType:  wechat.AppTypeMobile,
					BotProtection:  dummyBotProtection,
				},
				{
					Identification: config.AuthenticationFlowIdentificationPasskey,
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
                    "const": "username"
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
                "alias": {
                    "const": "google",
                    "type": "string"
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
                "bot_protection",
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
                "bot_protection",
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
                    "const": "passkey"
                }
            },
            "required": [
                "identification",
                "bot_protection",
                "assertion_response"
            ]
        }
    ],
    "type": "object"
}
`)
	})
}
