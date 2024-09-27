package usage

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	goredis "github.com/go-redis/redis/v8"
	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/lib/config"
)

func TestComputeResetTime(t *testing.T) {
	Convey("compute reset time of quota", t, func() {
		test := func(now string, day string, month string) {
			t, _ := time.Parse(time.RFC3339, now)
			So(
				ComputeResetTime(t, config.UsageLimitPeriodDay).Format(time.RFC3339),
				ShouldEqual,
				day,
			)
			So(
				ComputeResetTime(t, config.UsageLimitPeriodMonth).Format(time.RFC3339),
				ShouldEqual,
				month,
			)
		}
		test("2009-11-10T15:00:00Z", "2009-11-11T00:00:00Z", "2009-12-01T00:00:00Z")
		test("2009-02-28T23:00:00Z", "2009-03-01T00:00:00Z", "2009-03-01T00:00:00Z")
	})
}

const testKey = "rate-limit"
const epoch = 1257894000000

func TestLimitReserve(t *testing.T) {
	s := miniredis.RunT(t)

	Convey("reserve tokens", t, func() {
		ctx := context.Background()
		s.FlushAll()

		cli := goredis.NewClient(&goredis.Options{Addr: s.Addr()})
		conn := cli.Conn(ctx)

		now := time.UnixMilli(epoch)
		s.SetTime(now)

		forward := func(period time.Duration) {
			newNow := now.Add(period)
			s.SetTime(newNow)
			s.FastForward(newNow.Sub(now))
			now = newNow
		}

		quota := 10
		resetTime := time.UnixMilli(epoch).Add(12 * time.Hour)

		pass, tokens, err := reserve(ctx, conn, testKey, 1, quota, resetTime)
		So(err, ShouldBeNil)
		So(pass, ShouldBeTrue)
		So(tokens, ShouldEqual, 9)

		pass, tokens, err = reserve(ctx, conn, testKey, 8, quota, resetTime)
		So(err, ShouldBeNil)
		So(pass, ShouldBeTrue)
		So(tokens, ShouldEqual, 1)

		// Should not deduct tokens if the reservation attempt exceed it.
		pass, tokens, err = reserve(ctx, conn, testKey, 2, quota, resetTime)
		So(err, ShouldBeNil)
		So(pass, ShouldBeFalse)
		So(tokens, ShouldEqual, 1)

		pass, tokens, err = reserve(ctx, conn, testKey, 1, quota, resetTime)
		So(err, ShouldBeNil)
		So(pass, ShouldBeTrue)
		So(tokens, ShouldEqual, 0)

		pass, tokens, err = reserve(ctx, conn, testKey, 1, quota, resetTime)
		So(err, ShouldBeNil)
		So(pass, ShouldBeFalse)
		So(tokens, ShouldEqual, 0)

		forward(6 * time.Hour)

		pass, tokens, err = reserve(ctx, conn, testKey, 1, quota, resetTime)
		So(err, ShouldBeNil)
		So(pass, ShouldBeFalse)
		So(tokens, ShouldEqual, 0)

		forward(6 * time.Hour)

		pass, tokens, err = reserve(ctx, conn, testKey, 1, quota, resetTime)
		So(err, ShouldBeNil)
		So(pass, ShouldBeTrue)
		So(tokens, ShouldEqual, 9)
	})
}
