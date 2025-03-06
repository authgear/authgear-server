package redis

import (
	"bytes"
	"context"
	"log"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	goredis "github.com/redis/go-redis/v9"
	. "github.com/smartystreets/goconvey/convey"
)

var t_expired = time.Date(2025, 3, 5, 0, 0, 0, 0, time.UTC)
var t_now = time.Date(2025, 3, 5, 1, 0, 0, 0, time.UTC)
var t_future = time.Date(2025, 3, 5, 2, 0, 0, 0, time.UTC)

func testCleanUpNonExpiringKeysSetupFixture(ctx context.Context, redisClient *goredis.Client) {
	var err error

	set_WITH_EXPIRE := func(key string) {
		_, err = redisClient.Set(ctx, key, "not-important", 10*time.Second).Result()
		So(err, ShouldBeNil)
	}

	set_NO_EXPIRE := func(key string) {
		_, err = redisClient.Set(ctx, key, "not-important", 0).Result()
		So(err, ShouldBeNil)
	}
	hset := func(hashKey string, key string, expire time.Time) {
		b, err := expire.MarshalText()
		So(err, ShouldBeNil)
		_, err = redisClient.HSet(ctx, hashKey, key, b).Result()
		So(err, ShouldBeNil)
	}

	// These represent old non-expiring keys.
	set_NO_EXPIRE("app:accounts:access-events:old")
	set_NO_EXPIRE("app:accounts:failed-attempts:purpose:old")
	set_NO_EXPIRE("app:accounts:lockout:old")

	// These represent new expiring keys.
	set_WITH_EXPIRE("app:accounts:access-events:new")
	set_WITH_EXPIRE("app:accounts:failed-attempts:purpose:new")
	set_WITH_EXPIRE("app:accounts:lockout:new")

	hset("app:accounts:session-list:all-expired", "expired", t_expired)
	hset("app:accounts:session-list:some-expired", "expired", t_expired)
	hset("app:accounts:session-list:some-expired", "not-expired", t_future)
	hset("app:accounts:session-list:no-expired", "not-expired", t_future)

	hset("app:accounts:offline-grant-list:all-expired", "expired", t_expired)
	hset("app:accounts:offline-grant-list:some-expired", "expired", t_expired)
	hset("app:accounts:offline-grant-list:some-expired", "not-expired", t_future)
	hset("app:accounts:offline-grant-list:no-expired", "not-expired", t_future)
}

func TestCleanUpNonExpiringKeys(t *testing.T) {
	Convey("CleanUpNonExpiringKeys", t, func() {
		memoryRedis := miniredis.RunT(t)
		ctx := context.Background()
		redisClient := goredis.NewClient(&goredis.Options{Addr: memoryRedis.Addr()})
		stdout := &bytes.Buffer{}
		stderr := &bytes.Buffer{}
		logger := log.New(stderr, "", 0)
		testCleanUpNonExpiringKeysSetupFixture(ctx, redisClient)

		Convey("validate SCAN count", func() {
			err := CleanUpNonExpiringKeys(ctx, redisClient, stdout, logger, CleanUpNonExpiringKeysOptions{
				ScanCountString:  "a",
				KeyPattern:       "failed-attempts",
				ExpirationString: "1234s",
				DryRun:           true,
				Now:              t_now,
			})
			So(err, ShouldBeError, "SCAN count must be an integer: a")
		})

		Convey("validate expiration syntax", func() {
			err := CleanUpNonExpiringKeys(ctx, redisClient, stdout, logger, CleanUpNonExpiringKeysOptions{
				ScanCountString:  "10",
				KeyPattern:       "failed-attempts",
				ExpirationString: "a",
				DryRun:           true,
				Now:              t_now,
			})
			So(err, ShouldBeError, "expiration must be a valid Go duration literal: a")
		})

		Convey("validate expiration value", func() {
			err := CleanUpNonExpiringKeys(ctx, redisClient, stdout, logger, CleanUpNonExpiringKeysOptions{
				ScanCountString:  "10",
				KeyPattern:       "failed-attempts",
				ExpirationString: "-1s",
				DryRun:           true,
				Now:              t_now,
			})
			So(err, ShouldBeError, "expiration cannot be less than 0: -1s")
		})

		Convey("validate key pattern", func() {
			err := CleanUpNonExpiringKeys(ctx, redisClient, stdout, logger, CleanUpNonExpiringKeysOptions{
				ScanCountString:  "10",
				KeyPattern:       "a",
				ExpirationString: "1234s",
				DryRun:           true,
				Now:              t_now,
			})
			So(err, ShouldBeError, "unsupported key patttern: a")
		})

		Convey("validate now", func() {
			err := CleanUpNonExpiringKeys(ctx, redisClient, stdout, logger, CleanUpNonExpiringKeysOptions{
				ScanCountString:  "10",
				KeyPattern:       "session-list",
				ExpirationString: "1234s",
				DryRun:           true,
				Now:              time.Time{},
			})
			So(err, ShouldBeError, "now cannot be zero")
		})

		Convey("session-list dry-run=true", func() {
			err := CleanUpNonExpiringKeys(ctx, redisClient, stdout, logger, CleanUpNonExpiringKeysOptions{
				ScanCountString:  "10",
				KeyPattern:       "session-list",
				ExpirationString: "1234s",
				DryRun:           true,
				Now:              t_now,
			})
			So(err, ShouldBeNil)
			So(stderr.String(), ShouldEqual, `SCAN 0 app:*:session-list:* scanned_total=3 expired_total=0
(dry-run) EXPIRE app:accounts:session-list:all-expired 1234
done scanned_total=3 expired_total=1
`)
			So(stdout.String(), ShouldEqual, `app:accounts:session-list:all-expired
`)
		})

		Convey("offline-grant-list dry-run=true", func() {
			err := CleanUpNonExpiringKeys(ctx, redisClient, stdout, logger, CleanUpNonExpiringKeysOptions{
				ScanCountString:  "10",
				KeyPattern:       "offline-grant-list",
				ExpirationString: "1234s",
				DryRun:           true,
				Now:              t_now,
			})
			So(err, ShouldBeNil)
			So(stderr.String(), ShouldEqual, `SCAN 0 app:*:offline-grant-list:* scanned_total=3 expired_total=0
(dry-run) EXPIRE app:accounts:offline-grant-list:all-expired 1234
done scanned_total=3 expired_total=1
`)
			So(stdout.String(), ShouldEqual, `app:accounts:offline-grant-list:all-expired
`)
		})

		Convey("failed-attempts dry-run=true", func() {
			err := CleanUpNonExpiringKeys(ctx, redisClient, stdout, logger, CleanUpNonExpiringKeysOptions{
				ScanCountString:  "10",
				KeyPattern:       "failed-attempts",
				ExpirationString: "1234s",
				DryRun:           true,
				Now:              t_now,
			})
			So(err, ShouldBeNil)
			So(stderr.String(), ShouldEqual, `SCAN 0 app:*:failed-attempts:* scanned_total=2 expired_total=0
(dry-run) EXPIRE app:accounts:failed-attempts:purpose:old 1234
done scanned_total=2 expired_total=1
`)
			So(stdout.String(), ShouldEqual, `app:accounts:failed-attempts:purpose:old
`)
		})

		Convey("access-events dry-run=true", func() {
			err := CleanUpNonExpiringKeys(ctx, redisClient, stdout, logger, CleanUpNonExpiringKeysOptions{
				ScanCountString:  "10",
				KeyPattern:       "access-events",
				ExpirationString: "1234s",
				DryRun:           true,
				Now:              t_now,
			})
			So(err, ShouldBeNil)
			So(stderr.String(), ShouldEqual, `SCAN 0 app:*:access-events:* scanned_total=2 expired_total=0
(dry-run) EXPIRE app:accounts:access-events:old 1234
done scanned_total=2 expired_total=1
`)
			So(stdout.String(), ShouldEqual, `app:accounts:access-events:old
`)
		})

		Convey("lockout dry-run=true", func() {
			err := CleanUpNonExpiringKeys(ctx, redisClient, stdout, logger, CleanUpNonExpiringKeysOptions{
				ScanCountString:  "10",
				KeyPattern:       "lockout",
				ExpirationString: "1234s",
				DryRun:           true,
				Now:              t_now,
			})
			So(err, ShouldBeNil)
			So(stderr.String(), ShouldEqual, `SCAN 0 app:*:lockout:* scanned_total=2 expired_total=0
(dry-run) EXPIRE app:accounts:lockout:old 1234
done scanned_total=2 expired_total=1
`)
			So(stdout.String(), ShouldEqual, `app:accounts:lockout:old
`)
		})

		Convey("session-list dry-run=false", func() {
			err := CleanUpNonExpiringKeys(ctx, redisClient, stdout, logger, CleanUpNonExpiringKeysOptions{
				ScanCountString:  "10",
				KeyPattern:       "session-list",
				ExpirationString: "1234s",
				DryRun:           false,
				Now:              t_now,
			})
			So(err, ShouldBeNil)
			So(stderr.String(), ShouldEqual, `SCAN 0 app:*:session-list:* scanned_total=3 expired_total=0
EXPIRE app:accounts:session-list:all-expired 1234
done scanned_total=3 expired_total=1
`)
			So(stdout.String(), ShouldEqual, `app:accounts:session-list:all-expired
`)
		})

		Convey("offline-grant-list dry-run=false", func() {
			err := CleanUpNonExpiringKeys(ctx, redisClient, stdout, logger, CleanUpNonExpiringKeysOptions{
				ScanCountString:  "10",
				KeyPattern:       "offline-grant-list",
				ExpirationString: "1234s",
				DryRun:           false,
				Now:              t_now,
			})
			So(err, ShouldBeNil)
			So(stderr.String(), ShouldEqual, `SCAN 0 app:*:offline-grant-list:* scanned_total=3 expired_total=0
EXPIRE app:accounts:offline-grant-list:all-expired 1234
done scanned_total=3 expired_total=1
`)
			So(stdout.String(), ShouldEqual, `app:accounts:offline-grant-list:all-expired
`)
		})

		Convey("failed-attempts dry-run=false", func() {
			err := CleanUpNonExpiringKeys(ctx, redisClient, stdout, logger, CleanUpNonExpiringKeysOptions{
				ScanCountString:  "10",
				KeyPattern:       "failed-attempts",
				ExpirationString: "1234s",
				DryRun:           false,
				Now:              t_now,
			})
			So(err, ShouldBeNil)
			So(stderr.String(), ShouldEqual, `SCAN 0 app:*:failed-attempts:* scanned_total=2 expired_total=0
EXPIRE app:accounts:failed-attempts:purpose:old 1234
done scanned_total=2 expired_total=1
`)
			So(stdout.String(), ShouldEqual, `app:accounts:failed-attempts:purpose:old
`)
		})

		Convey("access-events dry-run=false", func() {
			err := CleanUpNonExpiringKeys(ctx, redisClient, stdout, logger, CleanUpNonExpiringKeysOptions{
				ScanCountString:  "10",
				KeyPattern:       "access-events",
				ExpirationString: "1234s",
				DryRun:           false,
				Now:              t_now,
			})
			So(err, ShouldBeNil)
			So(stderr.String(), ShouldEqual, `SCAN 0 app:*:access-events:* scanned_total=2 expired_total=0
EXPIRE app:accounts:access-events:old 1234
done scanned_total=2 expired_total=1
`)
			So(stdout.String(), ShouldEqual, `app:accounts:access-events:old
`)
		})

		Convey("lockout dry-run=false", func() {
			err := CleanUpNonExpiringKeys(ctx, redisClient, stdout, logger, CleanUpNonExpiringKeysOptions{
				ScanCountString:  "10",
				KeyPattern:       "lockout",
				ExpirationString: "1234s",
				DryRun:           false,
				Now:              t_now,
			})
			So(err, ShouldBeNil)
			So(stderr.String(), ShouldEqual, `SCAN 0 app:*:lockout:* scanned_total=2 expired_total=0
EXPIRE app:accounts:lockout:old 1234
done scanned_total=2 expired_total=1
`)
			So(stdout.String(), ShouldEqual, `app:accounts:lockout:old
`)
		})
	})
}
