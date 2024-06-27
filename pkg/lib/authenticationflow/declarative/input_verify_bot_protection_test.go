package declarative

import (
	"encoding/json"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestInputSchemaBotProtectionVerification(t *testing.T) {
	Convey("InputSchemaBotProtectionVerification", t, func() {
		test := func(s *InputSchemaBotProtectionVerification, expected string) {
			b := s.SchemaBuilder()
			bytes, err := json.Marshal(b)
			So(err, ShouldBeNil)
			So(string(bytes), ShouldEqualJSON, expected)
		}

		test(&InputSchemaBotProtectionVerification{}, `
{
    "additionalProperties": false,
    "properties": {
        "bot_protection": {
            "additionalProperties": false,
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
        "bot_protection"
    ],
    "type": "object"
}
`)
	})
}
