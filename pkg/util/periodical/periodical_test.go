package periodical_test

import (
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/periodical"
)

func TestPeriodicalArgumentParser(t *testing.T) {
	Convey("ParsePeriodicalArgumentParser", t, func() {

		parser := periodical.ArgumentParser{
			Clock: clock.NewMockClockAt("2006-01-02T15:04:05Z"),
		}

		var periodicalType periodical.Type
		var date *time.Time
		var err error

		periodicalType, date, err = parser.Parse("this-hour")
		So(periodicalType, ShouldEqual, periodical.Hourly)
		So(*date, ShouldResemble, time.Date(2006, 1, 2, 15, 0, 0, 0, time.UTC))
		So(err, ShouldBeNil)

		periodicalType, date, err = parser.Parse("today")
		So(periodicalType, ShouldEqual, periodical.Daily)
		So(*date, ShouldResemble, time.Date(2006, 1, 2, 0, 0, 0, 0, time.UTC))
		So(err, ShouldBeNil)

		periodicalType, date, err = parser.Parse("this-week")
		So(periodicalType, ShouldEqual, periodical.Weekly)
		So(*date, ShouldResemble, time.Date(2006, 1, 2, 0, 0, 0, 0, time.UTC))
		So(err, ShouldBeNil)

		periodicalType, date, err = parser.Parse("this-month")
		So(periodicalType, ShouldEqual, periodical.Monthly)
		So(*date, ShouldResemble, time.Date(2006, 1, 1, 0, 0, 0, 0, time.UTC))
		So(err, ShouldBeNil)

		periodicalType, date, err = parser.Parse("last-hour")
		So(periodicalType, ShouldEqual, periodical.Hourly)
		So(*date, ShouldResemble, time.Date(2006, 1, 2, 14, 0, 0, 0, time.UTC))
		So(err, ShouldBeNil)

		periodicalType, date, err = parser.Parse("yesterday")
		So(periodicalType, ShouldEqual, periodical.Daily)
		So(*date, ShouldResemble, time.Date(2006, 1, 1, 0, 0, 0, 0, time.UTC))
		So(err, ShouldBeNil)

		periodicalType, date, err = parser.Parse("last-week")
		So(periodicalType, ShouldEqual, periodical.Weekly)
		So(*date, ShouldResemble, time.Date(2005, 12, 26, 0, 0, 0, 0, time.UTC))
		So(err, ShouldBeNil)

		periodicalType, date, err = parser.Parse("last-month")
		So(periodicalType, ShouldEqual, periodical.Monthly)
		So(*date, ShouldResemble, time.Date(2005, 12, 1, 0, 0, 0, 0, time.UTC))
		So(err, ShouldBeNil)

		periodicalType, date, err = parser.Parse("2007-10")
		So(periodicalType, ShouldEqual, periodical.Monthly)
		So(*date, ShouldResemble, time.Date(2007, 10, 1, 0, 0, 0, 0, time.UTC))
		So(err, ShouldBeNil)

		periodicalType, date, err = parser.Parse("2007-10-24T02")
		So(periodicalType, ShouldEqual, periodical.Hourly)
		So(*date, ShouldResemble, time.Date(2007, 10, 24, 2, 0, 0, 0, time.UTC))
		So(err, ShouldBeNil)

		periodicalType, date, err = parser.Parse("2007-10-24")
		So(periodicalType, ShouldEqual, periodical.Daily)
		So(*date, ShouldResemble, time.Date(2007, 10, 24, 0, 0, 0, 0, time.UTC))
		So(err, ShouldBeNil)

		periodicalType, date, err = parser.Parse("2006-W37")
		So(periodicalType, ShouldEqual, periodical.Weekly)
		So(*date, ShouldResemble, time.Date(2006, 9, 11, 0, 0, 0, 0, time.UTC))
		So(err, ShouldBeNil)

		_, _, err = parser.Parse("2006W37")
		So(err, ShouldBeError, periodical.ErrInvalidPeriodical)

		_, _, err = parser.Parse("2021-W53")
		So(err, ShouldBeError, "invalid week: 2021W53 not in [1, 52]")

	})
}
