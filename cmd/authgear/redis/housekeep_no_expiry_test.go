package redis

import (
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func TestHousekeepTaskLockout(t *testing.T) {
	Convey("lockout", t, func() {
		f := newHousekeepFixture(t)

		// Deleted: no expiry. The record is a hash and the lock is a string, see pkg/lib/lockout/record.go.
		f.hsetRaw("app:accounts:lockout:AccountAuthentication:old", "total", "3")
		f.set("app:accounts:lockout:AccountAuthentication:old:lock:global", unixMilliString(t_expired), 0)

		// Left alone: has an expiry.
		f.hsetRaw("app:accounts:lockout:AccountAuthentication:new", "total", "3")
		f.expire("app:accounts:lockout:AccountAuthentication:new", time.Hour)
		f.set("app:accounts:lockout:AccountAuthentication:new:lock:global", unixMilliString(t_future), time.Hour)
		// Left alone: not a lockout key, though the name contains lockout.
		f.set("app:accounts:session:lockout:a", "1", 0)

		stdout, _ := f.run(false, housekeepTaskLockout)
		So(stdout, ShouldEqual, `deleted app:accounts:lockout:AccountAuthentication:old
deleted app:accounts:lockout:AccountAuthentication:old:lock:global
`)
		So(f.memoryRedis.Keys(), ShouldResemble, []string{
			"app:accounts:lockout:AccountAuthentication:new",
			"app:accounts:lockout:AccountAuthentication:new:lock:global",
			"app:accounts:session:lockout:a",
		})
	})
}

func TestHousekeepTaskFailedAttempts(t *testing.T) {
	Convey("failed-attempts", t, func() {
		f := newHousekeepFixture(t)

		// Deleted: no expiry.
		f.set("app:accounts:failed-attempts:oob-otp:+85200000000", "1", 0)

		// Left alone: has an expiry.
		f.set("app:accounts:failed-attempts:oob-otp:+85211111111", "1", time.Hour)
		// Left alone: not a failed attempts key, though the name contains failed-attempts.
		f.set("app:accounts:session:failed-attempts:a", "1", 0)

		stdout, _ := f.run(false, housekeepTaskFailedAttempts)
		So(stdout, ShouldEqual, "deleted app:accounts:failed-attempts:oob-otp:+85200000000\n")
		So(f.memoryRedis.Keys(), ShouldResemble, []string{
			"app:accounts:failed-attempts:oob-otp:+85211111111",
			"app:accounts:session:failed-attempts:a",
		})
	})
}
