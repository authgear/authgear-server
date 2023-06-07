package lockout

import (
	"context"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/alicebob/miniredis/v2"
	goredis "github.com/go-redis/redis/v8"
)

type testEntry struct {
	time     string
	attempts int

	expectedIsSucess    bool
	expectedLockedUntil *time.Time
}

type testConfig struct {
	historyDuration string
	maxAttempts     int
	minDuration     string
	maxDuration     string
	backoffFactor   float64
	entries         []testEntry
}

const testKey = "lockouttest"
const epoch = 1257894000

func TestMakeAttempt(t *testing.T) {
	s := miniredis.RunT(t)

	test := func(name string, cfg *testConfig) {
		Convey(name, func() {
			ctx := context.Background()
			s.FlushAll()

			cli := goredis.NewClient(&goredis.Options{Addr: s.Addr()})
			conn := cli.Conn(ctx)

			historyDuration, _ := time.ParseDuration(cfg.historyDuration)
			maxAttempts := cfg.maxAttempts
			minDuration, _ := time.ParseDuration(cfg.minDuration)
			maxDuration, _ := time.ParseDuration(cfg.maxDuration)
			backoffFactor := cfg.backoffFactor

			now := time.Unix(epoch, 0)
			for _, e := range cfg.entries {
				t, _ := time.ParseDuration(e.time)
				newNow := time.Unix(epoch, 0).Add(t)
				s.SetTime(newNow)
				s.FastForward(newNow.Sub(now))
				now = newNow

				result, err := makeAttempt(ctx, conn, testKey,
					historyDuration, maxAttempts, minDuration, maxDuration, backoffFactor, e.attempts)
				So(err, ShouldBeNil)
				So(result.IsSuccess, ShouldEqual, e.expectedIsSucess)
				So(result.LockedUntil, ShouldResemble, e.expectedLockedUntil)
			}
		})
	}

	Convey("Lockout", t, func() {
		test("makeAttempt", &testConfig{
			historyDuration: "300s",
			maxAttempts:     3,
			minDuration:     "10s",
			maxDuration:     "50s",
			backoffFactor:   2,
			entries: []testEntry{
				{time: "0s", attempts: 0, expectedIsSucess: true, expectedLockedUntil: nil},
				{time: "1s", attempts: 1, expectedIsSucess: true, expectedLockedUntil: nil},
				{time: "2s", attempts: 1, expectedIsSucess: true, expectedLockedUntil: nil},
				// The third attempt is still success
				{time: "3s", attempts: 1, expectedIsSucess: true, expectedLockedUntil: makeUnixTime(epoch + 3 + 10)},
				// The forth attempt is failed
				{time: "4s", attempts: 1, expectedIsSucess: false, expectedLockedUntil: makeUnixTime(epoch + 3 + 10)},
				// Lock again with min duration * 2
				{time: "13s", attempts: 1, expectedIsSucess: true, expectedLockedUntil: makeUnixTime(epoch + 13 + 20)},
				{time: "14s", attempts: 1, expectedIsSucess: false, expectedLockedUntil: makeUnixTime(epoch + 13 + 20)},
				// Lock again with min duration * 2 * 2
				{time: "33s", attempts: 1, expectedIsSucess: true, expectedLockedUntil: makeUnixTime(epoch + 33 + 40)},
				{time: "34s", attempts: 1, expectedIsSucess: false, expectedLockedUntil: makeUnixTime(epoch + 33 + 40)},
				// Lock again with min duration * 2 * 2 * 2, capped at max duration
				{time: "73s", attempts: 1, expectedIsSucess: true, expectedLockedUntil: makeUnixTime(epoch + 73 + 50)},
				{time: "74s", attempts: 1, expectedIsSucess: false, expectedLockedUntil: makeUnixTime(epoch + 73 + 50)},
				// Resets after history duration passed
				{time: "373s", attempts: 1, expectedIsSucess: true, expectedLockedUntil: nil},
				{time: "373s", attempts: 1, expectedIsSucess: true, expectedLockedUntil: nil},
				{time: "373s", attempts: 1, expectedIsSucess: true, expectedLockedUntil: makeUnixTime(epoch + 373 + 10)},
			},
		})
	})
}

func makeUnixTime(s int64) *time.Time {
	t := time.Unix(s, 0)
	return &t
}
