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
		So(IsValidUXMode(UXModeMobileApp), ShouldBeTrue)
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
			{[]string{"/A/B"}, "/a/b", false},
		}

		for _, c := range cases {
			So(f(c.urls, c.callbackURL) == nil, ShouldEqual, c.valid)
		}
	})
}
