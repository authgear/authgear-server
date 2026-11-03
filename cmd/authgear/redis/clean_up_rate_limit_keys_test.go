package redis

import (
	"bytes"
	"context"
	"log"
	"strconv"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	goredis "github.com/redis/go-redis/v9"
	. "github.com/smartystreets/goconvey/convey"
)

func TestCleanUpRateLimitKeys(t *testing.T) {
	Convey("CleanUpRateLimitKeys", t, func() {
		memoryRedis := miniredis.RunT(t)
		memoryRedis.SetTime(t_now)
		ctx := context.Background()
		redisClient := goredis.NewClient(&goredis.Options{Addr: memoryRedis.Addr()})
		stdout := &bytes.Buffer{}
		stderr := &bytes.Buffer{}
		logger := log.New(stderr, "", 0)

		set := func(key string, value string) {
			_, err := redisClient.Set(ctx, key, value, 0).Result()
			So(err, ShouldBeNil)
		}
		ms := func(t time.Time) string {
			return strconv.FormatInt(t.UnixMilli(), 10)
		}

		set("rate-limit:MessagingSMSPerTarget:+85200000000", ms(t_expired))
		set("app:accounts:rate-limit:SignupPerIP:127.0.0.1", ms(t_expired))
		set("app:accounts:rate-limit:OAuthTokenPerIP:127.0.0.1", ms(t_future))
		set("app:accounts:rate-limit:AuthflowAPIPerIP:127.0.0.1", "not-a-timestamp")
		_, err := redisClient.HSet(ctx, "app:accounts:rate-limit:VerifyPasswordPerIP:127.0.0.1", "field", "value").Result()
		So(err, ShouldBeNil)
		// Not rate limit keys, though they match the SCAN pattern.
		set("app:accounts:session:rate-limit:a", ms(t_expired))
		set("app:accounts:access-events:a", ms(t_expired))

		Convey("validate SCAN count", func() {
			err := CleanUpRateLimitKeys(ctx, redisClient, stdout, logger, CleanUpRateLimitKeysOptions{
				ScanCountString: "a",
				DryRun:          true,
				Now:             t_now,
			})
			So(err, ShouldBeError, "SCAN count must be an integer: a")
		})

		Convey("validate now", func() {
			err := CleanUpRateLimitKeys(ctx, redisClient, stdout, logger, CleanUpRateLimitKeysOptions{
				ScanCountString: "10",
				DryRun:          true,
				Now:             time.Time{},
			})
			So(err, ShouldBeError, "now cannot be zero")
		})

		Convey("dry-run=true", func() {
			err := CleanUpRateLimitKeys(ctx, redisClient, stdout, logger, CleanUpRateLimitKeysOptions{
				ScanCountString: "10",
				DryRun:          true,
				Now:             t_now,
			})
			So(err, ShouldBeNil)
			So(stdout.String(), ShouldEqual, `(dry-run) expiry-fixed app:accounts:rate-limit:OAuthTokenPerIP:127.0.0.1
(dry-run) deleted app:accounts:rate-limit:SignupPerIP:127.0.0.1
(dry-run) deleted rate-limit:MessagingSMSPerTarget:+85200000000
`)
			So(stderr.String(), ShouldEqual, `(dry-run) SCAN 0 *rate-limit:* scanned_total=5 deleted_total=2 expiry_fixed_total=1 skipped_total=2
(dry-run) done scanned_total=5 deleted_total=2 expiry_fixed_total=1 skipped_total=2
`)
			So(memoryRedis.Keys(), ShouldHaveLength, 7)
			So(memoryRedis.TTL("app:accounts:rate-limit:OAuthTokenPerIP:127.0.0.1"), ShouldEqual, 0)
		})

		Convey("dry-run=false", func() {
			err := CleanUpRateLimitKeys(ctx, redisClient, stdout, logger, CleanUpRateLimitKeysOptions{
				ScanCountString: "10",
				DryRun:          false,
				Now:             t_now,
			})
			So(err, ShouldBeNil)
			So(stdout.String(), ShouldEqual, `expiry-fixed app:accounts:rate-limit:OAuthTokenPerIP:127.0.0.1
deleted app:accounts:rate-limit:SignupPerIP:127.0.0.1
deleted rate-limit:MessagingSMSPerTarget:+85200000000
`)
			So(stderr.String(), ShouldEqual, `SCAN 0 *rate-limit:* scanned_total=5 deleted_total=2 expiry_fixed_total=1 skipped_total=2
done scanned_total=5 deleted_total=2 expiry_fixed_total=1 skipped_total=2
`)
			So(memoryRedis.Keys(), ShouldResemble, []string{
				"app:accounts:access-events:a",
				"app:accounts:rate-limit:AuthflowAPIPerIP:127.0.0.1",
				"app:accounts:rate-limit:OAuthTokenPerIP:127.0.0.1",
				"app:accounts:rate-limit:VerifyPasswordPerIP:127.0.0.1",
				"app:accounts:session:rate-limit:a",
			})
			So(memoryRedis.TTL("app:accounts:rate-limit:OAuthTokenPerIP:127.0.0.1"), ShouldEqual, t_future.Sub(t_now))
		})
	})
}
