package sso

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestIsValidUXmode(t *testing.T) {
	Convey("Test IsValidUXMode", t, func() {
		So(IsValidUXMode(""), ShouldBeFalse)
		So(IsValidUXMode("nonsense"), ShouldBeFalse)

		So(IsValidUXMode(UXModeWebRedirect), ShouldBeTrue)
		So(IsValidUXMode(UXModeWebPopup), ShouldBeTrue)
		So(IsValidUXMode(UXModeIOS), ShouldBeTrue)
		So(IsValidUXMode(UXModeAndroid), ShouldBeTrue)
	})
}

func TestIsValidOnUserDuplicate(t *testing.T) {
	Convey("Test IsValidOnUserDuplicate", t, func() {
		So(IsValidOnUserDuplicate(""), ShouldBeFalse)
		So(IsValidOnUserDuplicate("nonsense"), ShouldBeFalse)

		So(IsValidOnUserDuplicate(OnUserDuplicateDefault), ShouldBeTrue)
		So(IsValidOnUserDuplicate(OnUserDuplicateAbort), ShouldBeTrue)
		So(IsValidOnUserDuplicate(OnUserDuplicateMerge), ShouldBeTrue)
		So(IsValidOnUserDuplicate(OnUserDuplicateCreate), ShouldBeTrue)
	})
}
