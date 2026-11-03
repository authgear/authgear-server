package redis

import (
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func TestHousekeepTaskRateLimit(t *testing.T) {
	Convey("rate-limit", t, func() {
		f := newHousekeepFixture(t)
		ttlToFull := t_future.Sub(t_now)

		// Deleted: the bucket is full again.
		f.set("rate-limit:MessagingSMSPerTarget:+85200000000", unixMilliString(t_expired), 0)
		f.set("app:accounts:rate-limit:SignupPerIP:full", unixMilliString(t_expired), 0)
		// Written before timestamps were made integers.
		f.set("app:accounts:rate-limit:SignupPerIP:full-float", unixMilliString(t_expired)+".5", 0)

		// Expiry fixed: the bucket is not full yet, and the key has no expiry or one far too late.
		f.set("app:accounts:rate-limit:OAuthTokenPerIP:no-expiry", unixMilliString(t_future), 0)
		f.set("app:accounts:rate-limit:OAuthTokenPerIP:wrong-expiry", unixMilliString(t_future), 0)
		f.expire("app:accounts:rate-limit:OAuthTokenPerIP:wrong-expiry", 1000*24*time.Hour)

		// Left alone: the expiry is right, or off by less than the tolerance.
		f.set("app:accounts:rate-limit:OAuthTokenPerIP:right-expiry", unixMilliString(t_future), ttlToFull)
		f.set("app:accounts:rate-limit:OAuthTokenPerIP:within-tolerance", unixMilliString(t_future), ttlToFull+30*time.Second)
		f.set("app:accounts:rate-limit:OAuthTokenPerIP:earlier-expiry", unixMilliString(t_future), ttlToFull/2)
		// Left alone: not a value the application writes.
		f.set("app:accounts:rate-limit:AuthflowAPIPerIP:not-a-timestamp", "not-a-timestamp", 0)
		f.hsetRaw("app:accounts:rate-limit:VerifyPasswordPerIP:not-a-string", "field", unixMilliString(t_expired))
		// Left alone: not a rate limit key, though the name contains rate-limit.
		f.set("app:accounts:session:rate-limit:a", unixMilliString(t_expired), 0)

		Convey("dry-run=false", func() {
			stdout, _ := f.run(false, housekeepTaskRateLimit)
			So(stdout, ShouldEqual, `expiry-fixed app:accounts:rate-limit:OAuthTokenPerIP:no-expiry
expiry-fixed app:accounts:rate-limit:OAuthTokenPerIP:wrong-expiry
deleted app:accounts:rate-limit:SignupPerIP:full
deleted app:accounts:rate-limit:SignupPerIP:full-float
deleted rate-limit:MessagingSMSPerTarget:+85200000000
`)
			So(f.memoryRedis.Keys(), ShouldResemble, []string{
				"app:accounts:rate-limit:AuthflowAPIPerIP:not-a-timestamp",
				"app:accounts:rate-limit:OAuthTokenPerIP:earlier-expiry",
				"app:accounts:rate-limit:OAuthTokenPerIP:no-expiry",
				"app:accounts:rate-limit:OAuthTokenPerIP:right-expiry",
				"app:accounts:rate-limit:OAuthTokenPerIP:within-tolerance",
				"app:accounts:rate-limit:OAuthTokenPerIP:wrong-expiry",
				"app:accounts:rate-limit:VerifyPasswordPerIP:not-a-string",
				"app:accounts:session:rate-limit:a",
			})
			So(f.memoryRedis.TTL("app:accounts:rate-limit:OAuthTokenPerIP:no-expiry"), ShouldEqual, ttlToFull)
			So(f.memoryRedis.TTL("app:accounts:rate-limit:OAuthTokenPerIP:wrong-expiry"), ShouldEqual, ttlToFull)
			So(f.memoryRedis.TTL("app:accounts:rate-limit:OAuthTokenPerIP:within-tolerance"), ShouldEqual, ttlToFull+30*time.Second)
			So(f.memoryRedis.TTL("app:accounts:rate-limit:OAuthTokenPerIP:earlier-expiry"), ShouldEqual, ttlToFull/2)
		})
	})

	Convey("planRateLimitKey", t, func() {
		future := unixMilliString(t_future)
		msToFull := t_future.Sub(t_now).Milliseconds()

		test := func(value string, pttlMs int64, expected string) {
			result, _ := planRateLimitKey(value, pttlMs, t_now)
			So(result, ShouldEqual, expected)
		}

		test(unixMilliString(t_expired), -1, housekeepResultDeleted)
		test(unixMilliString(t_expired), msToFull, housekeepResultDeleted)
		test(unixMilliString(t_expired)+".5", -1, housekeepResultDeleted)

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
