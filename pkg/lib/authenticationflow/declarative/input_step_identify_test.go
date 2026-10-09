package declarative

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/oauthrelyingparty/wechat"
	"github.com/authgear/authgear-server/pkg/util/validation"
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
					Identification: model.AuthenticationFlowIdentificationEmail,
					BotProtection:  dummyBotProtection,
				},
				{
					Identification: model.AuthenticationFlowIdentificationPhone,
					BotProtection:  dummyBotProtection,
				},
				{
					Identification: model.AuthenticationFlowIdentificationUsername,
					BotProtection:  dummyBotProtection,
				},
				{
					Identification: model.AuthenticationFlowIdentificationOAuth,
					Alias:          "google",
					BotProtection:  dummyBotProtection,
				},
				{
					Identification: model.AuthenticationFlowIdentificationOAuth,
					Alias:          "wechat_mobile",
					WechatAppType:  wechat.AppTypeMobile,
					BotProtection:  dummyBotProtection,
				},
				{
					Identification: model.AuthenticationFlowIdentificationPasskey,
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

func TestInputSchemaStepIdentifySelectAccount(t *testing.T) {
	Convey("InputSchemaStepIdentify select_account", t, func() {
		test := func(s *InputSchemaStepIdentify, expected string) {
			b := s.SchemaBuilder()
			bytes, err := json.Marshal(b)
			So(err, ShouldBeNil)
			So(string(bytes), ShouldEqualJSON, expected)
		}

		// select_account's "index" is Const-validated against its own
		// position in the full Options slice, not the position among
		// select_account entries specifically.
		test(&InputSchemaStepIdentify{
			Options: []IdentificationOption{
				{
					Identification: model.AuthenticationFlowIdentificationEmail,
				},
				{
					Identification: model.AuthenticationFlowIdentificationSelectAccount,
					DisplayName:    "user@example.com",
				},
			},
		}, `
{
    "type": "object",
    "oneOf": [
        {
            "properties": {
                "identification": { "const": "email" },
                "login_id": { "type": "string" }
            },
            "required": ["identification", "login_id"]
        },
        {
            "properties": {
                "identification": { "const": "select_account" },
                "index": { "type": "integer", "const": 1 }
            },
            "required": ["identification", "index"]
        }
    ]
}
`)
	})
}

func TestInputSchemaStepIdentifyForms(t *testing.T) {
	Convey("InputSchemaStepIdentify with forms", t, func() {
		cfg := newTestFormConfig()
		requiredForm := newTestForm(userProfileField("/given_name", true))
		optionalForm := newTestForm(userProfileField("/nickname", false))

		// Several options per identification, as OAuth expands to one option per provider.
		s := &InputSchemaStepIdentify{
			Config: cfg,
			Options: []IdentificationOption{
				{Identification: model.AuthenticationFlowIdentificationUsername},
				{Identification: model.AuthenticationFlowIdentificationEmail},
				{Identification: model.AuthenticationFlowIdentificationPhone},
				{Identification: model.AuthenticationFlowIdentificationOAuth, Alias: "google"},
				{Identification: model.AuthenticationFlowIdentificationOAuth, Alias: "github"},
			},
			Forms: map[model.AuthenticationFlowIdentification]*config.AuthenticationFlowForm{
				model.AuthenticationFlowIdentificationUsername: requiredForm,
				model.AuthenticationFlowIdentificationPhone:    optionalForm,
			},
		}

		parse := func(raw string) (*InputStepIdentify, error) {
			input, err := s.MakeInput(context.Background(), json.RawMessage(raw))
			if err != nil {
				return nil, err
			}
			return input.(*InputStepIdentify), nil
		}

		Convey("fields are accepted on the branch with a form", func() {
			input, err := parse(`{
				"identification": "username",
				"login_id": "johndoe",
				"fields": [{"key": "/given_name", "value": "John"}]
			}`)
			So(err, ShouldBeNil)
			fields, submitted := input.GetIdentifyFormFields()
			So(submitted, ShouldBeTrue)
			So(fields, ShouldResemble, []FormFieldInput{{Key: "/given_name", Value: "John"}})
		})

		Convey("required fields are required", func() {
			_, err := parse(`{"identification": "username", "login_id": "johndoe"}`)
			So(err, ShouldNotBeNil)
			_, err = parse(`{"identification": "username", "login_id": "johndoe", "fields": []}`)
			So(err, ShouldNotBeNil)
		})

		Convey("optional fields may be omitted", func() {
			input, err := parse(`{"identification": "phone", "login_id": "+85298765432"}`)
			So(err, ShouldBeNil)
			fields, submitted := input.GetIdentifyFormFields()
			So(fields, ShouldBeEmpty)
			So(submitted, ShouldBeTrue)
		})

		Convey("fields are rejected on a branch without a form", func() {
			_, err := parse(`{
				"identification": "email",
				"login_id": "john@example.com",
				"fields": [{"key": "/given_name", "value": "John"}]
			}`)
			So(err, ShouldNotBeNil)
			// Rejected on every option that has no form, whatever the identification.
			_, err = parse(`{
				"identification": "oauth",
				"alias": "github",
				"redirect_uri": "https://example.com/callback",
				"fields": []
			}`)
			So(err, ShouldNotBeNil)
		})

		Convey("the fields of another branch are rejected", func() {
			_, err := parse(`{
				"identification": "phone",
				"login_id": "+85298765432",
				"fields": [{"key": "/given_name", "value": "John"}]
			}`)
			So(err, ShouldNotBeNil)
		})

		Convey("constraints are validated", func() {
			_, err := parse(`{
				"identification": "username",
				"login_id": "johndoe",
				"fields": [{"key": "/given_name", "value": "J"}]
			}`)
			So(err, ShouldNotBeNil)
			var aErr *validation.AggregatedError
			So(errors.As(err, &aErr), ShouldBeTrue)
			So(aErr.Errors[0].Keyword, ShouldEqual, "minLength")
			So(aErr.Errors[0].Location, ShouldEqual, "/fields/0/value")
		})

		Convey("a schema without forms is not changed", func() {
			s.Forms = nil
			input, err := parse(`{"identification": "email", "login_id": "john@example.com"}`)
			So(err, ShouldBeNil)
			So(input.Fields, ShouldBeEmpty)
		})
	})
}
