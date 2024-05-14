package analytic_test

import (
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/lib/analytic"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/periodical"
	"github.com/authgear/authgear-server/pkg/util/timeutil"
)

func TestChartService(t *testing.T) {
	Convey("TestChartService", t, func() {
		clock := clock.NewMockClockAt("2006-01-02T15:04:05Z")
		svc := &analytic.ChartService{
			Clock: clock,
			AnalyticConfig: &config.AnalyticConfig{
				Epoch: timeutil.Date(time.Date(2005, 6, 15, 0, 0, 0, 0, time.UTC)),
			},
		}
		Convey("TestGetBoundedRange", func() {
			// test daily range
			var rangeFrom, rangeTo time.Time
			var err error
			Convey("should bounded by epoch", func() {
				rangeFrom, rangeTo, err = svc.GetBoundedRange(
					periodical.Daily,
					time.Date(2005, 1, 1, 0, 0, 0, 0, time.UTC),
					time.Date(2006, 1, 1, 0, 0, 0, 0, time.UTC),
				)
				So(rangeFrom, ShouldResemble, time.Date(2005, 6, 15, 0, 0, 0, 0, time.UTC))
				So(rangeTo, ShouldResemble, time.Date(2006, 1, 1, 0, 0, 0, 0, time.UTC))
				So(err, ShouldBeNil)

				rangeFrom, rangeTo, err = svc.GetBoundedRange(
					periodical.Weekly,
					time.Date(2005, 1, 1, 0, 0, 0, 0, time.UTC),
					time.Date(2006, 1, 1, 0, 0, 0, 0, time.UTC),
				)
				So(rangeFrom, ShouldResemble, time.Date(2005, 6, 13, 0, 0, 0, 0, time.UTC))
				So(rangeTo, ShouldResemble, time.Date(2005, 12, 26, 0, 0, 0, 0, time.UTC))
				So(err, ShouldBeNil)

				rangeFrom, rangeTo, err = svc.GetBoundedRange(
					periodical.Monthly,
					time.Date(2005, 1, 1, 0, 0, 0, 0, time.UTC),
					time.Date(2006, 1, 1, 0, 0, 0, 0, time.UTC),
				)
				So(rangeFrom, ShouldResemble, time.Date(2005, 6, 1, 0, 0, 0, 0, time.UTC))
				So(rangeTo, ShouldResemble, time.Date(2005, 12, 1, 0, 0, 0, 0, time.UTC))
				So(err, ShouldBeNil)
			})

			Convey("should bounded by yesterday", func() {
				rangeFrom, rangeTo, err = svc.GetBoundedRange(
					periodical.Daily,
					time.Date(2005, 7, 1, 0, 0, 0, 0, time.UTC),
					time.Date(2006, 1, 3, 0, 0, 0, 0, time.UTC),
				)
				So(rangeFrom, ShouldResemble, time.Date(2005, 7, 1, 0, 0, 0, 0, time.UTC))
				So(rangeTo, ShouldResemble, time.Date(2006, 1, 1, 0, 0, 0, 0, time.UTC))
				So(err, ShouldBeNil)

				rangeFrom, rangeTo, err = svc.GetBoundedRange(
					periodical.Weekly,
					time.Date(2005, 7, 1, 0, 0, 0, 0, time.UTC),
					time.Date(2006, 1, 3, 0, 0, 0, 0, time.UTC),
				)
				So(rangeFrom, ShouldResemble, time.Date(2005, 6, 27, 0, 0, 0, 0, time.UTC))
				So(rangeTo, ShouldResemble, time.Date(2005, 12, 26, 0, 0, 0, 0, time.UTC))
				So(err, ShouldBeNil)

				rangeFrom, rangeTo, err = svc.GetBoundedRange(
					periodical.Monthly,
					time.Date(2005, 7, 1, 0, 0, 0, 0, time.UTC),
					time.Date(2006, 1, 3, 0, 0, 0, 0, time.UTC),
				)
				So(rangeFrom, ShouldResemble, time.Date(2005, 7, 1, 0, 0, 0, 0, time.UTC))
				So(rangeTo, ShouldResemble, time.Date(2005, 12, 1, 0, 0, 0, 0, time.UTC))
				So(err, ShouldBeNil)
			})

			Convey("should be valid to use specific day", func() {
				rangeFrom, rangeTo, err = svc.GetBoundedRange(
					periodical.Daily,
					time.Date(2006, 1, 1, 0, 0, 0, 0, time.UTC),
					time.Date(2006, 1, 1, 0, 0, 0, 0, time.UTC),
				)
				So(rangeFrom, ShouldResemble, time.Date(2006, 1, 1, 0, 0, 0, 0, time.UTC))
				So(rangeTo, ShouldResemble, time.Date(2006, 1, 1, 0, 0, 0, 0, time.UTC))
				So(err, ShouldBeNil)

				rangeFrom, rangeTo, err = svc.GetBoundedRange(
					periodical.Weekly,
					time.Date(2006, 1, 1, 0, 0, 0, 0, time.UTC),
					time.Date(2006, 1, 1, 0, 0, 0, 0, time.UTC),
				)
				So(rangeFrom, ShouldResemble, time.Date(2005, 12, 26, 0, 0, 0, 0, time.UTC))
				So(rangeTo, ShouldResemble, time.Date(2005, 12, 26, 0, 0, 0, 0, time.UTC))
				So(err, ShouldBeNil)

				rangeFrom, rangeTo, err = svc.GetBoundedRange(
					periodical.Monthly,
					time.Date(2005, 12, 6, 0, 0, 0, 0, time.UTC),
					time.Date(2005, 12, 6, 0, 0, 0, 0, time.UTC),
				)
				So(rangeFrom, ShouldResemble, time.Date(2005, 12, 1, 0, 0, 0, 0, time.UTC))
				So(rangeTo, ShouldResemble, time.Date(2005, 12, 1, 0, 0, 0, 0, time.UTC))
				So(err, ShouldBeNil)
			})

			Convey("should be error if it is out of range", func() {
				rangeFrom, rangeTo, err = svc.GetBoundedRange(
					periodical.Daily,
					time.Date(2006, 1, 2, 0, 0, 0, 0, time.UTC),
					time.Date(2006, 1, 2, 0, 0, 0, 0, time.UTC),
				)
				So(err, ShouldBeError, "invalid range")

				rangeFrom, rangeTo, err = svc.GetBoundedRange(
					periodical.Weekly,
					time.Date(2006, 1, 2, 0, 0, 0, 0, time.UTC),
					time.Date(2006, 1, 2, 0, 0, 0, 0, time.UTC),
				)
				So(err, ShouldBeError, "invalid range")

				rangeFrom, rangeTo, err = svc.GetBoundedRange(
					periodical.Monthly,
					time.Date(2006, 1, 1, 0, 0, 0, 0, time.UTC),
					time.Date(2006, 1, 1, 0, 0, 0, 0, time.UTC),
				)
				So(err, ShouldBeError, "invalid range")
			})

			Convey("should adjust range to first day of month", func() {
				rangeFrom, rangeTo, err = svc.GetBoundedRange(
					periodical.Monthly,
					time.Date(2005, 6, 30, 0, 0, 0, 0, time.UTC),
					time.Date(2005, 9, 6, 0, 0, 0, 0, time.UTC),
				)
				So(rangeFrom, ShouldResemble, time.Date(2005, 6, 1, 0, 0, 0, 0, time.UTC))
				So(rangeTo, ShouldResemble, time.Date(2005, 9, 1, 0, 0, 0, 0, time.UTC))
				So(err, ShouldBeNil)
			})

			Convey("should bound by last month", func() {
				rangeFrom, rangeTo, err = svc.GetBoundedRange(
					periodical.Monthly,
					time.Date(2005, 6, 30, 0, 0, 0, 0, time.UTC),
					time.Date(2006, 1, 20, 0, 0, 0, 0, time.UTC),
				)
				So(rangeFrom, ShouldResemble, time.Date(2005, 6, 1, 0, 0, 0, 0, time.UTC))
				So(rangeTo, ShouldResemble, time.Date(2005, 12, 1, 0, 0, 0, 0, time.UTC))
				So(err, ShouldBeNil)
			})

			Convey("should be valid for last month", func() {
				rangeFrom, rangeTo, err = svc.GetBoundedRange(
					periodical.Monthly,
					time.Date(2005, 12, 15, 0, 0, 0, 0, time.UTC),
					time.Date(2006, 1, 1, 0, 0, 0, 0, time.UTC),
				)
				So(rangeFrom, ShouldResemble, time.Date(2005, 12, 1, 0, 0, 0, 0, time.UTC))
				So(rangeTo, ShouldResemble, time.Date(2005, 12, 1, 0, 0, 0, 0, time.UTC))
				So(err, ShouldBeNil)
			})

			Convey("should adjust range to monday", func() {
				rangeFrom, rangeTo, err = svc.GetBoundedRange(
					periodical.Weekly,
					time.Date(2005, 6, 30, 0, 0, 0, 0, time.UTC),
					time.Date(2005, 9, 6, 0, 0, 0, 0, time.UTC),
				)
				So(rangeFrom, ShouldResemble, time.Date(2005, 6, 27, 0, 0, 0, 0, time.UTC))
				So(rangeTo, ShouldResemble, time.Date(2005, 9, 5, 0, 0, 0, 0, time.UTC))
				So(err, ShouldBeNil)
			})

		})

	})
}
