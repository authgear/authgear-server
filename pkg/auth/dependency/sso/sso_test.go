package sso

import (
	"testing"

	"github.com/skygeario/skygear-server/pkg/core/config"

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

func TestIsAllowedOnUserDuplicate(t *testing.T) {
	Convey("Test IsAllowedOnUserDuplicate", t, func() {
		c := config.OAuthConfiguration{}
		f := IsAllowedOnUserDuplicate

		So(f(c, OnUserDuplicateAbort), ShouldBeTrue)
		So(f(c, OnUserDuplicateMerge), ShouldBeFalse)
		So(f(c, OnUserDuplicateCreate), ShouldBeFalse)

		c.OnUserDuplicateAllowMerge = true

		So(f(c, OnUserDuplicateAbort), ShouldBeTrue)
		So(f(c, OnUserDuplicateMerge), ShouldBeTrue)
		So(f(c, OnUserDuplicateCreate), ShouldBeFalse)

		c.OnUserDuplicateAllowCreate = true

		So(f(c, OnUserDuplicateAbort), ShouldBeTrue)
		So(f(c, OnUserDuplicateMerge), ShouldBeTrue)
		So(f(c, OnUserDuplicateCreate), ShouldBeTrue)

		c.OnUserDuplicateAllowMerge = false

		So(f(c, OnUserDuplicateAbort), ShouldBeTrue)
		So(f(c, OnUserDuplicateMerge), ShouldBeFalse)
		So(f(c, OnUserDuplicateCreate), ShouldBeTrue)
	})
}
