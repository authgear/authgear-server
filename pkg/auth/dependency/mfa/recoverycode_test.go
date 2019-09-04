package mfa

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestGenerateRandomRecoveryCode(t *testing.T) {
	Convey("GenerateRandomRecoveryCode", t, func() {
		code := GenerateRandomRecoveryCode()
		So(len(code), ShouldEqual, 10)
	})
}
