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

func TestIsAllowedOnUserDuplicate(t *testing.T) {
	Convey("Test IsAllowedOnUserDuplicate", t, func() {
		f := IsAllowedOnUserDuplicate

		merge := false
		create := false

		So(f(merge, create, OnUserDuplicateAbort), ShouldBeTrue)
		So(f(merge, create, OnUserDuplicateMerge), ShouldBeFalse)
		So(f(merge, create, OnUserDuplicateCreate), ShouldBeFalse)

		merge = true

		So(f(merge, create, OnUserDuplicateAbort), ShouldBeTrue)
		So(f(merge, create, OnUserDuplicateMerge), ShouldBeTrue)
		So(f(merge, create, OnUserDuplicateCreate), ShouldBeFalse)

		create = true

		So(f(merge, create, OnUserDuplicateAbort), ShouldBeTrue)
		So(f(merge, create, OnUserDuplicateMerge), ShouldBeTrue)
		So(f(merge, create, OnUserDuplicateCreate), ShouldBeTrue)

		merge = false

		So(f(merge, create, OnUserDuplicateAbort), ShouldBeTrue)
		So(f(merge, create, OnUserDuplicateMerge), ShouldBeFalse)
		So(f(merge, create, OnUserDuplicateCreate), ShouldBeTrue)
	})
}

func TestValidateCallbackURL(t *testing.T) {
	Convey("Test ValidateCallbackURL", t, func() {
		f := ValidateCallbackURL

		So(f(nil, ""), ShouldBeError, "missing callback URL")

		cases := []struct {
			urls        []string
			callbackURL string
			valid       bool
		}{
			{nil, "a", false},
			{[]string{}, "a", false},
			{[]string{"b"}, "a", false},
			{[]string{"a"}, "a", true},
			{[]string{"/a"}, "/a/b", true},
			{[]string{"/a/c"}, "/a/b", false},
		}

		for _, c := range cases {
			So(f(c.urls, c.callbackURL) == nil, ShouldEqual, c.valid)
		}
	})
}
