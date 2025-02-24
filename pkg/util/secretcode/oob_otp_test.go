package secretcode

import (
	"context"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestOOBOTPSecretCode(t *testing.T) {
	Convey("OOBOTPSecretCode", t, func() {
		Convey("CheckFormat", func() {
			ctx := context.Background()
			f := OOBOTPSecretCode.CheckFormat
			So(f(ctx, nil), ShouldBeNil)
			So(f(ctx, 0), ShouldBeNil)
			So(f(ctx, false), ShouldBeNil)
			So(f(ctx, ""), ShouldBeError, "unexpected OOB OTP code length: 0")
			So(f(ctx, "1234a6"), ShouldBeError, `unexpected OOB OTP code character at index 4: "a"`)
			So(f(ctx, "123456"), ShouldBeNil)
			So(f(ctx, " 123456 "), ShouldBeNil)
			So(f(ctx, " 123 456 "), ShouldBeError, "unexpected OOB OTP code length: 7")
		})
	})
}
