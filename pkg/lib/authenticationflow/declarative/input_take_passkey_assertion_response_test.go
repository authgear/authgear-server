package declarative

import (
	"encoding/json"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestInputSchemaTakePasskeyAssertionResponse(t *testing.T) {
	Convey("InputSchemaTakePasskeyAssertionResponse", t, func() {
		test := func(s *InputSchemaTakePasskeyAssertionResponse, expected string) {
			b := s.SchemaBuilder()
			bytes, err := json.Marshal(b)
			So(err, ShouldBeNil)
			So(string(bytes), ShouldEqualJSON, expected)
		}

		test(&InputSchemaTakePasskeyAssertionResponse{}, `
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
        }
    },
    "required": [
        "assertion_response"
    ],
    "type": "object"
}
`)
	})
}
