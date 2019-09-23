package mfa

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestGenerateRandomOOBCode(t *testing.T) {
	Convey("GenerateRandomOOBCode", t, func() {
		code := GenerateRandomOOBCode()
		So(len(code), ShouldEqual, 6)
	})
}
