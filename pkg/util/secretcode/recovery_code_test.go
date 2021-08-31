package secretcode

import (
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
			So(f(nil), ShouldBeNil)
			So(f(0), ShouldBeNil)
			So(f(false), ShouldBeNil)
			So(f(""), ShouldBeError, "unexpected recovery code length: 0")
			So(f("!@#$%^&*"), ShouldBeError, "invalid recovery code: invalid base32 character: !")
			So(f("abcde-12345"), ShouldBeNil)
			So(f(" abcde-12345 "), ShouldBeNil)
			So(f(" abcde - 12345 "), ShouldBeNil)
			So(f("abcde12345"), ShouldBeNil)
		})
	})
}
