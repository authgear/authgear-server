package mfa

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestGenerateRandomBearerToken(t *testing.T) {
	Convey("GenerateRandomBearerToken", t, func() {
		code := GenerateRandomBearerToken()
		So(len(code), ShouldEqual, 64)
	})
}
