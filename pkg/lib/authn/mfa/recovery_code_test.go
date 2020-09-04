package mfa_test

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/lib/authn/mfa"
)

func TestRecoveryCode(t *testing.T) {
	Convey("RecoveryCode", t, func() {
		Convey("Generate -> Format -> Normalize", func() {
			code := mfa.GenerateRecoveryCode()
			formatted := mfa.FormatRecoveryCode(code)

			normalized, err := mfa.NormalizeRecoveryCode(formatted)
			So(err, ShouldBeNil)
			So(normalized, ShouldEqual, code)
		})
	})
}
