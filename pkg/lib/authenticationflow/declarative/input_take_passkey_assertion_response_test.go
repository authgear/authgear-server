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
                    "type": "string"
                },
                "response": {
                    "properties": {
                        "authenticatorData": {
                            "type": "string"
                        },
                        "clientDataJSON": {
                            "type": "string"
                        },
                        "signature": {
                            "type": "string"
                        },
                        "userHandle": {
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
