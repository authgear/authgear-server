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
	})
}
