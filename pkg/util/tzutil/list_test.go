package tzutil

import (
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func TestList(t *testing.T) {
	Convey("List", t, func() {
		ref := time.Date(2006, 1, 2, 3, 4, 5, 0, time.UTC)
		timezones, err := List(ref)
		So(err, ShouldBeNil)

		hkIndex := -1
		tokyoIndex := -1
		for idx, timezone := range timezones {
			if timezone.Name == "Asia/Hong_Kong" {
				hkIndex = idx
			}
			if timezone.Name == "Asia/Tokyo" {
				tokyoIndex = idx
			}
		}

		// Run some properties tests on timezones.
		// Hong Kong should be found.
		So(hkIndex, ShouldBeGreaterThanOrEqualTo, 0)
		// Tokyo should be found.
		So(tokyoIndex, ShouldBeGreaterThanOrEqualTo, 0)
		// Hong Kong and Tokyo does not have daylight saving.
		So(timezones[hkIndex].FormattedOffset, ShouldEqual, "+08:00")
		So(timezones[tokyoIndex].FormattedOffset, ShouldEqual, "+09:00")
		// Hokg Kong should be ordered before Tokyo.
		So(hkIndex, ShouldBeLessThan, tokyoIndex)
	})
}
