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
	time        string
	attempts    int
	contributor string

	expectedIsSucess    bool
	expectedLockedUntil *time.Time

	fn func(ctx context.Context, conn *goredis.Conn)
}

type testConfig struct {
	historyDuration string
	maxAttempts     int
	minDuration     string
	maxDuration     string
	backoffFactor   float64
	isGlobal        bool
	entries         []testEntry
}

const testKey = "lockouttest"
const epoch = 1257894000

func TestLockout(t *testing.T) {
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
			isGlobal := cfg.isGlobal

			now := time.Unix(epoch, 0)
			for _, e := range cfg.entries {
				if e.fn != nil {
					e.fn(ctx, conn)
					continue
				}

				t, _ := time.ParseDuration(e.time)
				newNow := time.Unix(epoch, 0).Add(t)
				s.SetTime(newNow)
				s.FastForward(newNow.Sub(now))
				now = newNow

				result, err := makeAttempts(ctx, conn, testKey,
					historyDuration, maxAttempts, minDuration, maxDuration, backoffFactor, isGlobal, e.contributor, e.attempts)
				So(err, ShouldBeNil)
				So(result.IsSuccess, ShouldEqual, e.expectedIsSucess)
				So(result.LockedUntil, ShouldResemble, e.expectedLockedUntil)
			}
		})
	}

	Convey("Lockout", t, func() {
		test("makeAttempts", &testConfig{
			historyDuration: "300s",
			maxAttempts:     3,
			minDuration:     "10s",
			maxDuration:     "50s",
			backoffFactor:   2,
			isGlobal:        true,
			entries: []testEntry{
				{time: "0s", contributor: "127.0.0.1", attempts: 0, expectedIsSucess: true, expectedLockedUntil: nil},
				{time: "1s", contributor: "127.0.0.1", attempts: 1, expectedIsSucess: true, expectedLockedUntil: nil},
				{time: "2s", contributor: "127.0.0.1", attempts: 1, expectedIsSucess: true, expectedLockedUntil: nil},
				// Checking with 0 attempts should be success without locking
				{time: "3s", contributor: "127.0.0.1", attempts: 0, expectedIsSucess: true, expectedLockedUntil: nil},
				// The third attempt is still success, but lock was created
				{time: "3s", contributor: "127.0.0.1", attempts: 1, expectedIsSucess: true, expectedLockedUntil: makeUnixTime(epoch + 3 + 10)},
				// The forth attempt is failed
				{time: "4s", contributor: "127.0.0.1", attempts: 1, expectedIsSucess: false, expectedLockedUntil: makeUnixTime(epoch + 3 + 10)},
				// 0 Attempts should also fail
				{time: "4s", contributor: "127.0.0.1", attempts: 0, expectedIsSucess: false, expectedLockedUntil: makeUnixTime(epoch + 3 + 10)},
				// Lock again with min duration * 2
				{time: "13s", contributor: "127.0.0.1", attempts: 1, expectedIsSucess: true, expectedLockedUntil: makeUnixTime(epoch + 13 + 20)},
				{time: "14s", contributor: "127.0.0.1", attempts: 1, expectedIsSucess: false, expectedLockedUntil: makeUnixTime(epoch + 13 + 20)},
				// Lock again with min duration * 2 * 2
				{time: "33s", contributor: "127.0.0.1", attempts: 1, expectedIsSucess: true, expectedLockedUntil: makeUnixTime(epoch + 33 + 40)},
				{time: "34s", contributor: "127.0.0.1", attempts: 1, expectedIsSucess: false, expectedLockedUntil: makeUnixTime(epoch + 33 + 40)},
				// Lock again with min duration * 2 * 2 * 2, capped at max duration
				{time: "73s", contributor: "127.0.0.1", attempts: 1, expectedIsSucess: true, expectedLockedUntil: makeUnixTime(epoch + 73 + 50)},
				{time: "74s", contributor: "127.0.0.1", attempts: 1, expectedIsSucess: false, expectedLockedUntil: makeUnixTime(epoch + 73 + 50)},
				// Resets after history duration passed
				{time: "373s", contributor: "127.0.0.1", attempts: 1, expectedIsSucess: true, expectedLockedUntil: nil},
				{time: "373s", contributor: "127.0.0.1", attempts: 1, expectedIsSucess: true, expectedLockedUntil: nil},
				{time: "373s", contributor: "127.0.0.1", attempts: 1, expectedIsSucess: true, expectedLockedUntil: makeUnixTime(epoch + 373 + 10)},
			},
		})

		test("makeAttempts global lock", &testConfig{
			historyDuration: "300s",
			maxAttempts:     3,
			minDuration:     "10s",
			maxDuration:     "50s",
			backoffFactor:   2,
			isGlobal:        true,
			entries: []testEntry{
				{time: "1s", contributor: "127.0.0.1", attempts: 3, expectedIsSucess: true, expectedLockedUntil: makeUnixTime(epoch + 1 + 10)},
				// attempt of the same contributor is failed
				{time: "2s", contributor: "127.0.0.1", attempts: 1, expectedIsSucess: false, expectedLockedUntil: makeUnixTime(epoch + 1 + 10)},
				// attempt of the other contributor is failed
				{time: "2s", contributor: "127.0.0.2", attempts: 1, expectedIsSucess: false, expectedLockedUntil: makeUnixTime(epoch + 1 + 10)},
			},
		})

		test("makeAttempts contributor lock", &testConfig{
			historyDuration: "300s",
			maxAttempts:     3,
			minDuration:     "10s",
			maxDuration:     "50s",
			backoffFactor:   2,
			isGlobal:        false,
			entries: []testEntry{
				{time: "1s", contributor: "127.0.0.1", attempts: 3, expectedIsSucess: true, expectedLockedUntil: makeUnixTime(epoch + 1 + 10)},
				// attempt of the same contributor is failed
				{time: "2s", contributor: "127.0.0.1", attempts: 1, expectedIsSucess: false, expectedLockedUntil: makeUnixTime(epoch + 1 + 10)},
				// attempt of the other contributor is success
				{time: "2s", contributor: "127.0.0.2", attempts: 1, expectedIsSucess: true, expectedLockedUntil: nil},

				// contributor specific lock locked after 3 attempts
				{time: "5s", contributor: "127.0.0.2", attempts: 2, expectedIsSucess: true, expectedLockedUntil: makeUnixTime(epoch + 5 + 10)},
				{time: "5s", contributor: "127.0.0.2", attempts: 1, expectedIsSucess: false, expectedLockedUntil: makeUnixTime(epoch + 5 + 10)},

				// lock of 127.0.0.1 is not affected
				{time: "5s", contributor: "127.0.0.1", attempts: 1, expectedIsSucess: false, expectedLockedUntil: makeUnixTime(epoch + 1 + 10)},
			},
		})

		test("clearAttempts global lock", &testConfig{
			historyDuration: "300s",
			maxAttempts:     3,
			minDuration:     "10s",
			maxDuration:     "50s",
			backoffFactor:   2,
			isGlobal:        true,
			entries: []testEntry{
				{time: "0s", contributor: "127.0.0.1", attempts: 4, expectedIsSucess: true, expectedLockedUntil: makeUnixTime(epoch + 0 + 20)},
				{time: "0s", fn: func(ctx context.Context, conn *goredis.Conn) {
					err := clearAttempts(ctx, conn, testKey, 300*time.Second, "127.0.0.1")
					So(err, ShouldBeNil)
				}},
				// Clear attempts should not affect existing lock
				{time: "1s", contributor: "127.0.0.1", attempts: 1, expectedIsSucess: false, expectedLockedUntil: makeUnixTime(epoch + 0 + 20)},
				// The lock time should reset on next lock
				{time: "20s", contributor: "127.0.0.1", attempts: 3, expectedIsSucess: true, expectedLockedUntil: makeUnixTime(epoch + 20 + 10)},
			},
		})

		test("clearAttempts contributor lock", &testConfig{
			historyDuration: "300s",
			maxAttempts:     3,
			minDuration:     "10s",
			maxDuration:     "50s",
			backoffFactor:   2,
			isGlobal:        false,
			entries: []testEntry{
				{time: "0s", contributor: "127.0.0.1", attempts: 4, expectedIsSucess: true, expectedLockedUntil: makeUnixTime(epoch + 0 + 20)},
				{time: "0s", contributor: "127.0.0.2", attempts: 5, expectedIsSucess: true, expectedLockedUntil: makeUnixTime(epoch + 0 + 40)},
				{time: "0s", fn: func(ctx context.Context, conn *goredis.Conn) {
					err := clearAttempts(ctx, conn, testKey, 300*time.Second, "127.0.0.1")
					So(err, ShouldBeNil)
				}},
				// Clear attempts should not affect existing lock
				{time: "1s", contributor: "127.0.0.1", attempts: 1, expectedIsSucess: false, expectedLockedUntil: makeUnixTime(epoch + 0 + 20)},
				{time: "1s", contributor: "127.0.0.2", attempts: 1, expectedIsSucess: false, expectedLockedUntil: makeUnixTime(epoch + 0 + 40)},
				// The lock time of 127.0.0.1 should reset on next lock
				{time: "20s", contributor: "127.0.0.1", attempts: 3, expectedIsSucess: true, expectedLockedUntil: makeUnixTime(epoch + 20 + 10)},
				// The lock time of 127.0.0.2 should not reset on next lock
				{time: "40s", contributor: "127.0.0.2", attempts: 1, expectedIsSucess: true, expectedLockedUntil: makeUnixTime(epoch + 40 + 50)},
			},
		})
	})
}

func TestLockoutClearAttempts(t *testing.T) {
	s := miniredis.RunT(t)

	Convey("clearAttempts should set ttl if the key originally has no ttl set", t, func() {
		ctx := context.Background()
		cli := goredis.NewClient(&goredis.Options{Addr: s.Addr()})
		conn := cli.Conn(ctx)

		err := clearAttempts(ctx, conn, testKey, 300*time.Second, "127.0.0.1")
		So(err, ShouldBeNil)

		ttl, err := conn.TTL(ctx, testKey).Result()
		So(err, ShouldBeNil)
		So(ttl, ShouldBeGreaterThan, 295*time.Second)
		So(ttl, ShouldBeLessThanOrEqualTo, 300*time.Second)
	})

	Convey("clearAttempts should not set ttl if the key originally has ttl set", t, func() {
		ctx := context.Background()
		cli := goredis.NewClient(&goredis.Options{Addr: s.Addr()})
		conn := cli.Conn(ctx)

		_, err := conn.HSet(ctx, testKey, "127.0.0.1", "1").Result()
		So(err, ShouldBeNil)
		// The original ttl is 200s
		_, err = conn.Expire(ctx, testKey, 200*time.Second).Result()
		So(err, ShouldBeNil)

		// The proposed ttl is 300s
		err = clearAttempts(ctx, conn, testKey, 300*time.Second, "127.0.0.1")
		So(err, ShouldBeNil)

		// The actual ttl should still around 200s
		ttl, err := conn.TTL(ctx, testKey).Result()
		So(err, ShouldBeNil)
		So(ttl, ShouldBeGreaterThan, 195*time.Second)
		So(ttl, ShouldBeLessThanOrEqualTo, 200*time.Second)
	})
}

func makeUnixTime(s int64) *time.Time {
	t := time.Unix(s, 0)
	return &t
}
