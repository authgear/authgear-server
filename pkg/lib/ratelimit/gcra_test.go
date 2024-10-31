package ratelimit

import (
	"context"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/alicebob/miniredis/v2"
	goredis "github.com/redis/go-redis/v9"
)

type scheduleEntry struct {
	time string
	n    int

	ok        bool
	timeToAct int

	fn func()
}

type schedule struct {
	period  string
	burst   int
	entries []scheduleEntry
}

const testKey = "rate-limit"
const epoch = 1257894000000

func TestGCRA(t *testing.T) {
	s := miniredis.RunT(t)

	test := func(name string, sch *schedule) {
		Convey(name, func() {
			ctx := context.Background()
			s.FlushAll()

			cli := goredis.NewClient(&goredis.Options{Addr: s.Addr()})
			conn := cli.Conn()

			period, _ := time.ParseDuration(sch.period)
			burst := sch.burst

			now := time.UnixMilli(epoch).UTC()
			for _, e := range sch.entries {
				if e.fn != nil {
					e.fn()
					continue
				}

				t, _ := time.ParseDuration(e.time)
				newNow := time.UnixMilli(epoch).UTC().Add(t)
				s.SetTime(newNow)
				s.FastForward(newNow.Sub(now))
				now = newNow

				result, err := gcra(ctx, conn, testKey, period, burst, e.n)
				So(err, ShouldBeNil)
				So(result.IsConforming, ShouldEqual, e.ok)
				So((result.TimeToAct.UnixMilli()-epoch)/1000, ShouldEqual, e.timeToAct)
			}
		})
	}

	Convey("GCRA", t, func() {
		test("basic", &schedule{
			period: "20s",
			burst:  4,
			entries: []scheduleEntry{
				{time: "0s", n: 0, ok: true, timeToAct: -15},
				{time: "0s", n: 1, ok: true, timeToAct: -10},
				{time: "0s", n: 1, ok: true, timeToAct: -5},
				{time: "0s", n: 1, ok: true, timeToAct: 0},
				{time: "0s", n: 1, ok: true, timeToAct: 5}, // token exhausted, wait 5s

				{time: "1s", n: 1, ok: false, timeToAct: 5},
				{time: "2s", n: 1, ok: false, timeToAct: 5},
				{time: "3s", n: 1, ok: false, timeToAct: 5},
				{time: "4s", n: 1, ok: false, timeToAct: 5},
				{time: "5s", n: 1, ok: true, timeToAct: 10}, // 1 token refilled after 5s

				{time: "6s", n: 1, ok: false, timeToAct: 10},
				{time: "7s", n: 1, ok: false, timeToAct: 10},
				{time: "8s", n: 1, ok: false, timeToAct: 10},
				{time: "9s", n: 1, ok: false, timeToAct: 10},
				{time: "10s", n: 1, ok: true, timeToAct: 15}, // 1 token refilled after another 5s

				{time: "10s", n: 1, ok: false, timeToAct: 15},
				{time: "11s", n: 0, ok: true, timeToAct: 15}, // taking 0 tokens always succeed
				{time: "11s", n: 1, ok: false, timeToAct: 15},

				{time: "100s", n: 0, ok: true, timeToAct: 85}, // tokens are not accumulated past burst limits
				{time: "100s", n: 1, ok: true, timeToAct: 90},
				{time: "100s", n: 1, ok: true, timeToAct: 95},
				{time: "100s", n: 1, ok: true, timeToAct: 100},
				{time: "100s", n: 1, ok: true, timeToAct: 105},
				{time: "100s", n: 1, ok: false, timeToAct: 105},
				{time: "100s", n: 0, ok: true, timeToAct: 105},
			},
		})
		test("substained requests", &schedule{
			period: "20s",
			burst:  4,
			entries: []scheduleEntry{
				{time: "0s", n: 1, ok: true, timeToAct: -10}, // allow substained requests at 4/20s (i.e. 1/5s)
				{time: "5s", n: 1, ok: true, timeToAct: -5},
				{time: "10s", n: 1, ok: true, timeToAct: 0},
				{time: "15s", n: 1, ok: true, timeToAct: 5},
				{time: "20s", n: 1, ok: true, timeToAct: 10},
				{time: "25s", n: 1, ok: true, timeToAct: 15},
				{time: "30s", n: 1, ok: true, timeToAct: 20},
			},
		})
		test("burst requests", &schedule{
			period: "20s",
			burst:  4,
			entries: []scheduleEntry{
				{time: "0s", n: 5, ok: false, timeToAct: 5}, // allow burst requests within limit
				{time: "0s", n: 4, ok: true, timeToAct: 20},
				{time: "1s", n: 1, ok: false, timeToAct: 5},
				{time: "1s", n: 0, ok: true, timeToAct: 5},

				{time: "100s", n: 5, ok: false, timeToAct: 105},
				{time: "100s", n: 3, ok: true, timeToAct: 110},
				{time: "101s", n: 1, ok: true, timeToAct: 105},
				{time: "101s", n: 0, ok: true, timeToAct: 105},

				{time: "200s", n: 5, ok: false, timeToAct: 205},
				{time: "200s", n: 2, ok: true, timeToAct: 200},
				{time: "201s", n: 2, ok: true, timeToAct: 210},
				{time: "201s", n: 0, ok: true, timeToAct: 205},
			},
		})
		test("reservation & return", &schedule{
			period: "10s",
			burst:  2,
			entries: []scheduleEntry{
				{time: "0s", n: 0, ok: true, timeToAct: -5},
				{time: "0s", n: 1, ok: true, timeToAct: 0},
				{time: "0s", n: -1, ok: true, timeToAct: -5}, // return reserved tokens, back to original state

				{time: "0s", n: 1, ok: true, timeToAct: 0},
				{time: "3s", n: 1, ok: true, timeToAct: 5},
				{time: "6s", n: -1, ok: true, timeToAct: 0},

				{time: "8s", n: 1, ok: true, timeToAct: 8},
				{time: "9s", n: 1, ok: true, timeToAct: 13},
				{time: "10s", n: -1, ok: true, timeToAct: 8},
				{time: "11s", n: -1, ok: true, timeToAct: 3},
				{time: "20s", n: -1, ok: true, timeToAct: 10},
				{time: "21s", n: 1, ok: true, timeToAct: 21}, // ignore excessive returns (i.e. underflow)
				{time: "22s", n: 1, ok: true, timeToAct: 26},
				{time: "23s", n: 1, ok: false, timeToAct: 26},
			},
		})
		test("ignore old rate limit state", &schedule{
			period: "10s",
			burst:  2,
			entries: []scheduleEntry{
				{fn: func() {
					s.HSet(testKey, "test", "123")
				}},
				{time: "0s", n: 0, ok: true, timeToAct: -5},
				{time: "0s", n: 1, ok: true, timeToAct: 0},
				{time: "0s", n: -1, ok: true, timeToAct: -5},
			},
		})
	})
}
