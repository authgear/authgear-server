package user

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestAccountStatus(t *testing.T) {
	Convey("AccountStatus", t, func() {
		var normal AccountStatus
		var err error

		_, err = normal.Reenable()
		So(err, ShouldBeError, "invalid account status transition: normal -> normal")

		disabled, err := normal.Disable(nil)
		So(err, ShouldBeNil)
		So(disabled.Type(), ShouldEqual, AccountStatusTypeDisabled)

		_, err = disabled.Disable(nil)
		So(err, ShouldBeError, "invalid account status transition: disabled -> disabled")

		normalAgain, err := disabled.Reenable()
		So(err, ShouldBeNil)
		So(normalAgain.Type(), ShouldEqual, AccountStatusTypeNormal)
	})
}
