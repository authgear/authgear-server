package secretcode

import (
	"context"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestRecoveryCode(t *testing.T) {
	Convey("RecoveryCode", t, func() {
		Convey("Generate -> Format -> Normalize", func() {
			code := RecoveryCode.Generate()
			formatted := RecoveryCode.FormatForHuman(code)

			normalized, err := RecoveryCode.FormatForComparison(formatted)
			So(err, ShouldBeNil)
			So(normalized, ShouldEqual, code)
		})

		Convey("CheckFormat", func() {
			f := RecoveryCode.CheckFormat
			ctx := context.Background()
			So(f(ctx, nil), ShouldBeNil)
			So(f(ctx, 0), ShouldBeNil)
			So(f(ctx, false), ShouldBeNil)
			So(f(ctx, ""), ShouldBeError, "unexpected recovery code length: 0")
			So(f(ctx, "!@#$%^&*"), ShouldBeError, "invalid recovery code: invalid base32 character: !")
			So(f(ctx, "abcde-12345"), ShouldBeNil)
			So(f(ctx, " abcde-12345 "), ShouldBeNil)
			So(f(ctx, " abcde - 12345 "), ShouldBeNil)
			So(f(ctx, "abcde12345"), ShouldBeNil)
		})
	})
}
