package usage

import (
	"errors"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/api/model"
)

func TestStandingUsageLimitDetails(t *testing.T) {
	Convey("StandingUsageLimitDetails", t, func() {
		Convey("a real ErrStandingUsageLimitExceeded round-trips its usage name and quota", func() {
			err := ErrStandingUsageLimitExceeded(model.UsageNameOAuthClientCIMD, 20)
			name, quota, ok := StandingUsageLimitDetails(err)
			So(ok, ShouldBeTrue)
			So(name, ShouldEqual, model.UsageNameOAuthClientCIMD)
			So(quota, ShouldEqual, 20)
		})

		Convey("a different usage name and quota round-trip too, proving this isn't hardcoded to one call site", func() {
			err := ErrStandingUsageLimitExceeded(model.UsageNameOAuthClientDCR, 5)
			name, quota, ok := StandingUsageLimitDetails(err)
			So(ok, ShouldBeTrue)
			So(name, ShouldEqual, model.UsageNameOAuthClientDCR)
			So(quota, ShouldEqual, 5)
		})

		Convey("a plain infrastructure error: ok is false, name and quota are zero", func() {
			name, quota, ok := StandingUsageLimitDetails(errors.New("redis: connection refused"))
			So(ok, ShouldBeFalse)
			So(name, ShouldEqual, model.UsageName(""))
			So(quota, ShouldEqual, 0)
		})

		Convey("an APIError of an unrelated Kind: ok is false", func() {
			err := apierrors.NotFound.WithReason("SomeUnrelatedError").New("some other error entirely")
			_, _, ok := StandingUsageLimitDetails(err)
			So(ok, ShouldBeFalse)
		})

		Convey("nil: ok is false", func() {
			_, _, ok := StandingUsageLimitDetails(nil)
			So(ok, ShouldBeFalse)
		})
	})
}
