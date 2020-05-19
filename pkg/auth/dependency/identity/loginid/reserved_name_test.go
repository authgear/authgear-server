package loginid

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestReservedNameChecker(t *testing.T) {

	Convey("TestReservedNameChecker", t, func() {
		checker, _ := NewReservedNameChecker("../../../../../reserved_name.txt")

		var result bool
		var err error

		result, err = checker.IsReserved("is")
		So(err, ShouldBeNil)
		So(result, ShouldBeTrue)

		result, err = checker.IsReserved("mail")
		So(err, ShouldBeNil)
		So(result, ShouldBeTrue)

		result, err = checker.IsReserved("faseng")
		So(result, ShouldBeFalse)
		So(err, ShouldBeNil)
	})

}
