package redis

import (
	"strconv"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func TestHousekeepTaskRateLimit(t *testing.T) {
	Convey("planRateLimitKey", t, func() {
		ms := func(t time.Time) string {
			return strconv.FormatInt(t.UnixMilli(), 10)
		}
		future := ms(t_future)
		msToFull := t_future.Sub(t_now).Milliseconds()

		test := func(value string, pttlMs int64, expected string) {
			result, _ := planRateLimitKey(value, pttlMs, t_now)
			So(result, ShouldEqual, expected)
		}

		test(ms(t_expired), -1, housekeepResultDeleted)
		test(ms(t_expired), msToFull, housekeepResultDeleted)
		test(ms(t_expired)+".5", -1, housekeepResultDeleted)

		test(future, -1, housekeepResultExpiryFixed)
		// Before DEV-3995, the millisecond timestamp was given to EXPIREAT as seconds.
		// This is the PTTL of such a key, about 56,000 years, which overflows time.Duration.
		test(future, t_future.UnixMilli()*1000-t_now.UnixMilli(), housekeepResultExpiryFixed)
		test(future, msToFull+time.Minute.Milliseconds()+1, housekeepResultExpiryFixed)

		test(future, msToFull, "")
		test(future, msToFull+time.Minute.Milliseconds(), "")
		test(future, msToFull/2, "")
		test("not-a-timestamp", -1, "")
		test("", -1, "")
	})
}
