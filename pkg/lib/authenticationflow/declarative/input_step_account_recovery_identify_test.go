package declarative

import (
	"encoding/json"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/lib/config"
)

func TestInputSchemaStepAccountRecoveryIdentify(t *testing.T) {
	Convey("InputSchemaStepAccountRecoveryIdentify", t, func() {
		test := func(s *InputSchemaStepAccountRecoveryIdentify, expected string) {
			b := s.SchemaBuilder()
			bytes, err := json.Marshal(b)
			So(err, ShouldBeNil)
			So(string(bytes), ShouldEqualJSON, expected)
		}

		test(&InputSchemaStepAccountRecoveryIdentify{
			Options: []AccountRecoveryIdentificationOption{
				{
					Identification: config.AuthenticationFlowAccountRecoveryIdentificationEmail,
				},
				{
					Identification: config.AuthenticationFlowAccountRecoveryIdentificationPhone,
				},
			},
		}, `
{
    "oneOf": [
        {
            "properties": {
                "identification": {
                    "const": "email"
                },
                "login_id": {
                    "type": "string"
                }
            },
            "required": [
                "identification",
                "login_id"
            ]
        },
        {
            "properties": {
                "identification": {
                    "const": "phone"
                },
                "login_id": {
                    "type": "string"
                }
            },
            "required": [
                "identification",
                "login_id"
            ]
        }
    ],
    "type": "object"
}
`)
	})
}
