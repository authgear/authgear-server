package authinfo

import (
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func TestAuthInfo(t *testing.T) {
	Convey("return new AuthInfo", t, func() {
		info := NewAuthInfo()
		So(info.ID, ShouldNotBeEmpty)
	})
}

func TestDisabled(t *testing.T) {
	timeNow := func() time.Time {
		return time.Date(2017, 12, 2, 0, 0, 0, 0, time.UTC)
	}

	Convey("Test default AuthInfo is not disabled", t, func() {
		info := AuthInfo{}
		So(info.IsDisabled(timeNow()), ShouldBeFalse)
	})

	Convey("Test AuthInfo.IsDisabled", t, func() {
		info := AuthInfo{}

		Convey("default should be false", func() {
			So(info.IsDisabled(timeNow()), ShouldBeFalse)
		})

		Convey("should return true if Disabled=true", func() {
			info.Disabled = true
			So(info.IsDisabled(timeNow()), ShouldBeTrue)
		})

		Convey("should return true if DisabledExpiry is in the future", func() {
			info.Disabled = true
			expiry := timeNow().Add(time.Hour)
			info.DisabledExpiry = &expiry
			So(info.IsDisabled(timeNow()), ShouldBeTrue)
		})

		Convey("should return false if DisabledExpiry is in the past", func() {
			info.Disabled = true
			expiry := timeNow().Add(-1 * time.Hour)
			info.DisabledExpiry = &expiry
			So(info.IsDisabled(timeNow()), ShouldBeFalse)
		})
	})

	Convey("Test AuthInfo.RefreshDisabledStatus", t, func() {
		info := AuthInfo{}

		Convey("should set Disabled to false when expiry is in the past", func() {
			info.Disabled = true
			info.DisabledMessage = "some reason"
			expiry := timeNow().Add(-1 * time.Hour)
			info.DisabledExpiry = &expiry
			info.RefreshDisabledStatus(timeNow())
			So(info.Disabled, ShouldBeFalse)
			So(info.DisabledExpiry, ShouldBeNil)
			So(info.DisabledMessage, ShouldBeEmpty)
		})
	})
}
