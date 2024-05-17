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
	"type": "object",
	"required": ["query"],
	"properties": {
		"query": {
			"type": "string"
		}
	}
}
`)
	})
}
