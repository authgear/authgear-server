package declarative

import (
	"encoding/json"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

func TestInputSchemaTakeOOBOTPChannel(t *testing.T) {
	Convey("InputSchemaTakeOOBOTPChannel", t, func() {
		test := func(b validation.SchemaBuilder, expected string) {
			bytes, err := json.Marshal(b)
			So(err, ShouldBeNil)
			So(string(bytes), ShouldEqualJSON, expected)
		}

		test((&InputSchemaTakeOOBOTPChannel{
			Channels: []model.AuthenticatorOOBChannel{model.AuthenticatorOOBChannelEmail},
		}).SchemaBuilder(), `
{
    "type": "object",
    "required": [
        "channel"
    ],
    "properties": {
        "channel": {
            "type": "string",
            "enum": ["email"]
        }
    }
}
		`)
	})
}
