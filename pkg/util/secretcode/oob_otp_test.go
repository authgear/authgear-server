package secretcode

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestOOBOTPSecretCode(t *testing.T) {
	Convey("OOBOTPSecretCode", t, func() {
		Convey("CheckFormat", func() {
			f := OOBOTPSecretCode.CheckFormat
			So(f(nil), ShouldBeNil)
			So(f(0), ShouldBeNil)
			So(f(false), ShouldBeNil)
			So(f(""), ShouldBeError, "unexpected OOB OTP code length: 0")
			So(f("1234a6"), ShouldBeError, `unexpected OOB OTP code character at index 4: "a"`)
			So(f("123456"), ShouldBeNil)
		})
	})
}
