package timeutil

import (
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func TestLast30Days(t *testing.T) {
	Convey("Last30Days", t, func() {
		Convey("should return the correct range for a normal date", func() {
			now := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
			from, to := Last30Days(now)
			So(to, ShouldResemble, now)
			So(from, ShouldResemble, time.Date(2023, 12, 2, 0, 0, 0, 0, time.UTC))
		})

		Convey("should handle leap years", func() {
			now := time.Date(2024, 3, 15, 0, 0, 0, 0, time.UTC)
			from, to := Last30Days(now)
			So(to, ShouldResemble, now)
			So(from, ShouldResemble, time.Date(2024, 2, 14, 0, 0, 0, 0, time.UTC))
		})

		Convey("should handle month boundaries", func() {
			now := time.Date(2024, 5, 1, 0, 0, 0, 0, time.UTC)
			from, to := Last30Days(now)
			So(to, ShouldResemble, now)
			So(from, ShouldResemble, time.Date(2024, 4, 1, 0, 0, 0, 0, time.UTC))
		})
	})
}
