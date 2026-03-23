package service

import (
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
)

func testTime() time.Time {
	return time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
}

func TestNewDomain(t *testing.T) {
	Convey("newDomain", t, func() {
		Convey("derives apex domain via PSL for a public TLD", func() {
			d, err := newDomain("app1", "auth.example.com", testTime(), true)
			So(err, ShouldBeNil)
			So(d.ApexDomain, ShouldEqual, "example.com")
		})

		Convey("returns InvalidDomain error for an invalid domain name", func() {
			_, err := newDomain("app1", "notadomain", testTime(), true)
			So(err, ShouldNotBeNil)
			So(apierrors.IsKind(err, InvalidDomain), ShouldBeTrue)
		})
	})
}

func TestOverrideApexDomain(t *testing.T) {
	Convey("domain.overrideApexDomain", t, func() {
		Convey("accepts a valid parent domain", func() {
			d, err := newDomain("app1", "auth.admin.hanlun-lms-dev.pandawork.com", testTime(), true)
			So(err, ShouldBeNil)
			// PSL derives pandawork.com; override to a more specific parent
			err = d.overrideApexDomain("admin.hanlun-lms-dev.pandawork.com")
			So(err, ShouldBeNil)
			So(d.ApexDomain, ShouldEqual, "admin.hanlun-lms-dev.pandawork.com")
		})

		Convey("accepts when apex equals the domain itself", func() {
			d, err := newDomain("app1", "auth.example.com", testTime(), true)
			So(err, ShouldBeNil)
			err = d.overrideApexDomain("auth.example.com")
			So(err, ShouldBeNil)
			So(d.ApexDomain, ShouldEqual, "auth.example.com")
		})

		Convey("rejects a domain that is not a parent with InvalidApexDomain reason", func() {
			d, err := newDomain("app1", "auth.example.com", testTime(), true)
			So(err, ShouldBeNil)
			err = d.overrideApexDomain("other.com")
			So(err, ShouldNotBeNil)
			// Must be InvalidApexDomain (not InvalidDomain) so the frontend
			// routes the error to the verification domain field, not the domain field.
			So(apierrors.IsKind(err, InvalidApexDomain), ShouldBeTrue)
			So(apierrors.IsKind(err, InvalidDomain), ShouldBeFalse)
			So(err.Error(), ShouldContainSubstring, `expected a suffix of "auth.example.com"`)
			So(err.Error(), ShouldContainSubstring, `got "other.com"`)
		})

		Convey("rejects a partial label match that is not a proper DNS parent", func() {
			// "xample.com" is a suffix of "auth.example.com" as a string but not as a DNS label
			d, err := newDomain("app1", "auth.example.com", testTime(), true)
			So(err, ShouldBeNil)
			err = d.overrideApexDomain("xample.com")
			So(err, ShouldNotBeNil)
			So(apierrors.IsKind(err, InvalidApexDomain), ShouldBeTrue)
		})
	})
}
