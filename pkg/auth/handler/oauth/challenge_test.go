package oauth

import (
	"context"
	"strings"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/util/validation"
)

func TestChallengeRequestValidation(t *testing.T) {
	Convey("ChallengeRequest validation", t, func() {
		validate := func(body string) error {
			var payload ChallengeRequest
			ctx := context.Background()
			reader := strings.NewReader(body)
			err := ChallengeAPIRequestSchema.Validator().ParseWithMessage(ctx, reader, "msg", &payload)
			if err != nil {
				return err
			}
			err = validation.ValidateValueWithMessage(ctx, &payload, "msg")
			if err != nil {
				return err
			}
			return nil
		}

		test := func(body string, expected string) {
			err := validate(body)
			if expected == "" {
				So(err, ShouldBeNil)
			} else {
				So(err, ShouldBeError, expected)
			}
		}

		test(`{}`, `msg:
<root>: required
  map[actual:<nil> expected:[purpose] missing:[purpose]]`)

		test(`{
			"purpose": "foobar"
		}`, `msg:
/purpose: unknown challenge purpose`)

		test(`{
			"purpose": "biometric_request"
		}`, ``)
	})
}
