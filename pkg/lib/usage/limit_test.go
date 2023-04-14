package usage_test

import (
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/usage"
)

func TestComputeResetTime(t *testing.T) {
	Convey("compute reset time of quota", t, func() {
		test := func(now string, day string, month string) {
			t, _ := time.Parse(time.RFC3339, now)
			So(
				usage.ComputeResetTime(t, config.UsageLimitPeriodDay).Format(time.RFC3339),
				ShouldEqual,
				day,
			)
			So(
				usage.ComputeResetTime(t, config.UsageLimitPeriodMonth).Format(time.RFC3339),
				ShouldEqual,
				month,
			)
		}
		test("2009-11-10T15:00:00Z", "2009-11-11T00:00:00Z", "2009-12-01T00:00:00Z")
		test("2009-02-28T23:00:00Z", "2009-03-01T00:00:00Z", "2009-03-01T00:00:00Z")
	})
}
