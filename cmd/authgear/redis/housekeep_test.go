package redis

import (
	"bytes"
	"context"
	"log"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	goredis "github.com/redis/go-redis/v9"
	. "github.com/smartystreets/goconvey/convey"
)

func TestHousekeep(t *testing.T) {
	Convey("Housekeep", t, func() {
		memoryRedis := miniredis.RunT(t)
		memoryRedis.SetTime(t_now)
		ctx := context.Background()
		redisClient := goredis.NewClient(&goredis.Options{Addr: memoryRedis.Addr()})
		stdout := &bytes.Buffer{}
		stderr := &bytes.Buffer{}
		logger := log.New(stderr, "", 0)

		set := func(key string, value string, ttl time.Duration) {
			_, err := redisClient.Set(ctx, key, value, ttl).Result()
			So(err, ShouldBeNil)
		}
		xadd := func(key string) {
			_, err := redisClient.XAdd(ctx, &goredis.XAddArgs{Stream: key, Values: map[string]any{"access": "{}"}}).Result()
			So(err, ShouldBeNil)
		}
		ms := func(t time.Time) string {
			return strconv.FormatInt(t.UnixMilli(), 10)
		}
		run := func(dryRun bool) error {
			stdout.Reset()
			stderr.Reset()
			return Housekeep(ctx, redisClient, stdout, logger, HousekeepOptions{
				ScanCountString: "100",
				DryRun:          dryRun,
				Now:             t_now,
			})
		}

		// rate-limit
		set("rate-limit:MessagingSMSPerTarget:+85200000000", ms(t_expired), 0)
		set("app:accounts:rate-limit:SignupPerIP:127.0.0.1", ms(t_expired), 0)
		set("app:accounts:rate-limit:OAuthTokenPerIP:127.0.0.1", ms(t_future), 0)
		set("app:accounts:rate-limit:AuthflowAPIPerIP:127.0.0.1", "not-a-timestamp", 0)
		_, err := redisClient.HSet(ctx, "app:accounts:rate-limit:VerifyPasswordPerIP:127.0.0.1", "field", "value").Result()
		So(err, ShouldBeNil)
		// access-events
		xadd("app:accounts:access-events:session-a")
		set("app:accounts:access-events:not-a-stream", "value", 0)
		// lockout and failed-attempts
		set("app:accounts:lockout:old", "1", 0)
		set("app:accounts:lockout:new", "1", time.Hour)
		set("app:accounts:failed-attempts:purpose:old", "1", 0)
		set("app:accounts:failed-attempts:purpose:new", "1", time.Hour)
		// Keys that are not housekept, including session lists, which are left for now.
		set("app:accounts:session:rate-limit:a", ms(t_expired), 0)
		xadd("app:accounts:session:access-events:a")
		_, err = redisClient.HSet(ctx, "app:accounts:session-list:user", "session", t_expired.Format(time.RFC3339)).Result()
		So(err, ShouldBeNil)
		set("app:accounts:offline-grant:grant", "{}", 0)

		allKeys := memoryRedis.Keys()

		Convey("validate SCAN count", func() {
			err := Housekeep(ctx, redisClient, stdout, logger, HousekeepOptions{
				ScanCountString: "a",
				DryRun:          true,
				Now:             t_now,
			})
			So(err, ShouldBeError, "SCAN count must be an integer: a")
		})

		Convey("validate now", func() {
			err := Housekeep(ctx, redisClient, stdout, logger, HousekeepOptions{
				ScanCountString: "100",
				DryRun:          true,
				Now:             time.Time{},
			})
			So(err, ShouldBeError, "now cannot be zero")
		})

		Convey("dry-run=true changes nothing", func() {
			err := run(true)
			So(err, ShouldBeNil)
			So(stdout.String(), ShouldEqual, `(dry-run) deleted app:accounts:access-events:session-a
(dry-run) deleted app:accounts:failed-attempts:purpose:old
(dry-run) deleted app:accounts:lockout:old
(dry-run) expiry-fixed app:accounts:rate-limit:OAuthTokenPerIP:127.0.0.1
(dry-run) deleted app:accounts:rate-limit:SignupPerIP:127.0.0.1
(dry-run) deleted rate-limit:MessagingSMSPerTarget:+85200000000
`)
			So(stderr.String(), ShouldEqual, `(dry-run) SCAN 0 scanned_total=15 deleted=5 expiry_fixed=1 skipped=5
(dry-run) done rate-limit deleted=2 expiry_fixed=1 skipped=2
(dry-run) done access-events deleted=1 expiry_fixed=0 skipped=1
(dry-run) done lockout deleted=1 expiry_fixed=0 skipped=1
(dry-run) done failed-attempts deleted=1 expiry_fixed=0 skipped=1
(dry-run) done scanned_total=15 deleted=5 expiry_fixed=1 skipped=5
`)
			So(memoryRedis.Keys(), ShouldResemble, allKeys)
			So(memoryRedis.TTL("app:accounts:rate-limit:OAuthTokenPerIP:127.0.0.1"), ShouldEqual, 0)
		})

		Convey("dry-run=false", func() {
			err := run(false)
			So(err, ShouldBeNil)
			So(stdout.String(), ShouldEqual, `deleted app:accounts:access-events:session-a
deleted app:accounts:failed-attempts:purpose:old
deleted app:accounts:lockout:old
expiry-fixed app:accounts:rate-limit:OAuthTokenPerIP:127.0.0.1
deleted app:accounts:rate-limit:SignupPerIP:127.0.0.1
deleted rate-limit:MessagingSMSPerTarget:+85200000000
`)
			So(memoryRedis.Keys(), ShouldResemble, []string{
				"app:accounts:access-events:not-a-stream",
				"app:accounts:failed-attempts:purpose:new",
				"app:accounts:lockout:new",
				"app:accounts:offline-grant:grant",
				"app:accounts:rate-limit:AuthflowAPIPerIP:127.0.0.1",
				"app:accounts:rate-limit:OAuthTokenPerIP:127.0.0.1",
				"app:accounts:rate-limit:VerifyPasswordPerIP:127.0.0.1",
				"app:accounts:session-list:user",
				"app:accounts:session:access-events:a",
				"app:accounts:session:rate-limit:a",
			})
			So(memoryRedis.TTL("app:accounts:rate-limit:OAuthTokenPerIP:127.0.0.1"), ShouldEqual, t_future.Sub(t_now))

			Convey("a rerun changes nothing", func() {
				keys := memoryRedis.Keys()
				err := run(false)
				So(err, ShouldBeNil)
				So(stdout.String(), ShouldEqual, "")
				So(stderr.String(), ShouldContainSubstring, "done scanned_total=10 deleted=0 expiry_fixed=0 skipped=6\n")
				So(memoryRedis.Keys(), ShouldResemble, keys)
			})
		})

		Convey("handles the keyspace in batches", func() {
			err := Housekeep(ctx, redisClient, stdout, logger, HousekeepOptions{
				ScanCountString: "10",
				DryRun:          true,
				Now:             t_now,
			})
			So(err, ShouldBeNil)
			So(strings.Count(stderr.String(), "SCAN "), ShouldBeGreaterThan, 1)
			So(stderr.String(), ShouldContainSubstring, "done scanned_total=15 deleted=5 expiry_fixed=1 skipped=5\n")
		})
	})
}
