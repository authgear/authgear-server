package loginid

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestReservedNameChecker(t *testing.T) {

	Convey("TestReservedNameChecker", t, func() {
		checker := NewReservedNameChecker([]string{
			"is",
			"mail",
		})

		var result bool

		result = checker.IsReserved("is")
		So(result, ShouldBeTrue)

		result = checker.IsReserved("mail")
		So(result, ShouldBeTrue)

		result = checker.IsReserved("faseng")
		So(result, ShouldBeFalse)
	})

}
