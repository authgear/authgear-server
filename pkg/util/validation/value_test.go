package validation_test

import (
	"testing"

	"github.com/authgear/authgear-server/pkg/util/validation"

	. "github.com/smartystreets/goconvey/convey"
)

type structA struct {
	X structB
	Y string
}

type structB struct {
	Z int
}

func (s structA) Validate(ctx *validation.Context) {
	ctx.Child("x").Validate(s.X)
	if s.Y == "" {
		ctx.Child("y").EmitErrorMessage("y is required")
	}
}

func (s structB) Validate(ctx *validation.Context) {
	if s.Z < 12 {
		ctx.Child("z").EmitError("minimum", map[string]interface{}{"minimum": 12})
	}
}

func TestValueValidate(t *testing.T) {
	Convey("validate value", t, func() {
		err := validation.ValidateValue(&structA{})
		So(err, ShouldBeError, `invalid value:
/x/z: minimum
  map[minimum:12]
/y: y is required`)

		err = validation.ValidateValue(&structA{
			X: structB{Z: 10},
			Y: "test",
		})
		So(err, ShouldBeError, `invalid value:
/x/z: minimum
  map[minimum:12]`)

		err = validation.ValidateValue(&structA{
			X: structB{Z: 20},
			Y: "test",
		})
		So(err, ShouldBeNil)
	})
}
