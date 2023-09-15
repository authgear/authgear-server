package declarative

import (
	"encoding/json"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestInputSchemaPromptCreatePasskey(t *testing.T) {
	Convey("InputSchemaPromptCreatePasskey", t, func() {
		test := func(s *InputSchemaPromptCreatePasskey, expected string) {
			b := s.SchemaBuilder()
			bytes, err := json.Marshal(b)
			So(err, ShouldBeNil)
			So(string(bytes), ShouldEqualJSON, expected)
		}

		test(&InputSchemaPromptCreatePasskey{}, `
{
    "oneOf": [
        {
            "properties": {
                "creation_response": {
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
                                "attestationObject": {
                                    "format": "x_base64_url",
                                    "type": "string"
                                },
                                "clientDataJSON": {
                                    "format": "x_base64_url",
                                    "type": "string"
                                }
                            },
                            "required": [
                                "attestationObject",
                                "clientDataJSON"
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
                }
            },
            "required": [
                "creation_response"
            ],
            "type": "object"
        },
        {
            "properties": {
                "skip": {
                    "type": "boolean"
                }
            },
            "required": [
                "skip"
            ],
            "type": "object"
        }
    ],
    "type": "object"
}
`)
	})
}
