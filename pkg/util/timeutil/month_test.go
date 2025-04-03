package timeutil_test

import (
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/util/timeutil"
)

func TestPreviousMonth(t *testing.T) {
	Convey("PreviousMonth", t, func() {
		date := func(year int, month time.Month, day int) time.Time {
			return time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
		}

		test := func(t time.Time, year int, month int) {
			actualYear, actualMonth := timeutil.PreviousMonth(t)
			So(actualYear, ShouldEqual, year)
			So(int(actualMonth), ShouldEqual, month)
		}

		Convey("The previous month of 2025-01-31 is 2025-12", func() {
			test(date(2025, 1, 31), 2024, 12)
		})

		Convey("The previous month of 2025-03-31 is 2025-02", func() {
			test(date(2025, 3, 31), 2025, 2)
		})
	})
}
