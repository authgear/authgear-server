package validation_test

import (
	"context"
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

func (s structA) Validate(ctx context.Context, vctx *validation.Context) {
	vctx.Child("x").Validate(ctx, s.X)
	if s.Y == "" {
		vctx.Child("y").EmitErrorMessage("y is required")
	}
}

func (s structB) Validate(ctx context.Context, vctx *validation.Context) {
	if s.Z < 12 {
		vctx.Child("z").EmitError("minimum", map[string]interface{}{"minimum": 12})
	}
}

func TestValueValidate(t *testing.T) {
	ctx := context.Background()
	Convey("validate value", t, func() {
		err := validation.ValidateValue(ctx, &structA{})
		So(err, ShouldBeError, `invalid value:
/x/z: minimum
  map[minimum:12]
/y: y is required`)

		err = validation.ValidateValue(ctx, &structA{
			X: structB{Z: 10},
			Y: "test",
		})
		So(err, ShouldBeError, `invalid value:
/x/z: minimum
  map[minimum:12]`)

		err = validation.ValidateValue(ctx, &structA{
			X: structB{Z: 20},
			Y: "test",
		})
		So(err, ShouldBeNil)
	})
}
