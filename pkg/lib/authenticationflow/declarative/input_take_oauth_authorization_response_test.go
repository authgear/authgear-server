package declarative

import (
	"encoding/json"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestInputSchemaTakeOAuthAuthorizationResponse(t *testing.T) {
	Convey("InputSchemaTakeOAuthAuthorizationResponse", t, func() {
		test := func(s *InputSchemaTakeOAuthAuthorizationResponse, expected string) {
			b := s.SchemaBuilder()
			bytes, err := json.Marshal(b)
			So(err, ShouldBeNil)
			So(string(bytes), ShouldEqualJSON, expected)
		}

		test(&InputSchemaTakeOAuthAuthorizationResponse{}, `
{
    "oneOf": [
        {
            "properties": {
                "code": {
                    "type": "string"
                },
                "error": {
                    "type": "string"
                },
                "error_description": {
                    "type": "string"
                },
                "error_uri": {
                    "type": "string"
                }
            },
            "required": [
                "code"
            ],
            "type": "object"
        },
        {
            "required": [
                "error"
            ],
            "type": "object"
        }
    ],
    "type": "object"
}
		`)
	})
}
