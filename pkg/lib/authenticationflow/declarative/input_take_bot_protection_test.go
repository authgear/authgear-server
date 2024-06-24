package declarative

import (
	"encoding/json"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestInputSchemaBotProtectionVerification(t *testing.T) {
	Convey("InputSchemaBotProtectionVerification", t, func() {
		test := func(expected string) {
			b := NewInputTakeBotProtectionSchemaBuilder()
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
}
