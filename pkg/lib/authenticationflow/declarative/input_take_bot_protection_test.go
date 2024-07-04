package declarative

import (
	"encoding/json"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestInputTakeBotProtection(t *testing.T) {
	Convey("InputTakeBotProtectionBody", t, func() {
		test := func(expected string) {
			b := InputTakeBotProtectionBodySchemaBuilder
			bytes, err := json.Marshal(b)
			So(err, ShouldBeNil)
			So(string(bytes), ShouldEqualJSON, expected)
		}

		test(`
{
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
`)
	})
	Convey("InputTakeBotProtection", t, func() {
		test := func(expected string) {
			b := InputTakeBotProtectionSchemaBuilder
			bytes, err := json.Marshal(b)
			So(err, ShouldBeNil)
			So(string(bytes), ShouldEqualJSON, expected)
		}
		test(`
{
    "type": "object",
    "required": ["bot_protection"],
    "properties": {
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
    }
}
`)
	})
}
