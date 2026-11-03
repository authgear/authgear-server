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
		hset := func(key string, field string, expireAt time.Time) {
			b, err := expireAt.MarshalText()
			So(err, ShouldBeNil)
			_, err = redisClient.HSet(ctx, key, field, string(b)).Result()
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

		// session-list: the field is the session key.
		set("app:accounts:session:live", "{}", time.Hour)
		set("app:accounts:session:stale-expiry", "{}", time.Hour)
		hset("app:accounts:session-list:all-dead", "app:accounts:session:dead-1", t_expired)
		hset("app:accounts:session-list:mixed", "app:accounts:session:dead-2", t_expired)
		hset("app:accounts:session-list:mixed", "app:accounts:session:live", t_future)
		// The application writes the entry before the session key.
		hset("app:accounts:session-list:mixed", "app:accounts:session:being-created", t_future)
		// The session key still exists, so the entry is kept.
		hset("app:accounts:session-list:mixed", "app:accounts:session:stale-expiry", t_expired)
		_, err = redisClient.HSet(ctx, "app:accounts:session-list:unreadable", "app:accounts:session:dead-3", "not-a-time").Result()
		So(err, ShouldBeNil)
		set("app:accounts:session-list:not-a-hash", "value", 0)

		// offline-grant-list: the field is the bare offline grant ID.
		set("app:accounts:offline-grant:grant-live", "{}", time.Hour)
		hset("app:accounts:offline-grant-list:user-a", "grant-live", t_future)
		hset("app:accounts:offline-grant-list:user-a", "grant-dead-1", t_expired)
		hset("app:accounts:offline-grant-list:user-b", "grant-dead-2", t_expired)

		// Keys that are not housekept.
		set("app:accounts:session:rate-limit:a", ms(t_expired), 0)
		xadd("app:accounts:session:access-events:a")

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

		expectedOutput := `expiry-fixed app:accounts:rate-limit:OAuthTokenPerIP:127.0.0.1
deleted app:accounts:rate-limit:SignupPerIP:127.0.0.1
deleted rate-limit:MessagingSMSPerTarget:+85200000000
deleted app:accounts:access-events:session-a
deleted app:accounts:lockout:old
deleted app:accounts:failed-attempts:purpose:old
deleted app:accounts:session-list:all-dead
pruned app:accounts:session-list:mixed
pruned app:accounts:offline-grant-list:user-a
deleted app:accounts:offline-grant-list:user-b
`
		expectedSummary := `done rate-limit deleted=2 expiry_fixed=1 pruned=0 skipped=2 conflict=0
done access-events deleted=1 expiry_fixed=0 pruned=0 skipped=1 conflict=0
done lockout deleted=1 expiry_fixed=0 pruned=0 skipped=1 conflict=0
done failed-attempts deleted=1 expiry_fixed=0 pruned=0 skipped=1 conflict=0
done session-list deleted=1 expiry_fixed=0 pruned=1 skipped=2 conflict=0
done offline-grant-list deleted=1 expiry_fixed=0 pruned=1 skipped=0 conflict=0
`

		Convey("dry-run=true changes nothing", func() {
			err := run(true)
			So(err, ShouldBeNil)
			So(stdout.String(), ShouldEqual, prefixLines("(dry-run) ", expectedOutput))
			So(stderr.String(), ShouldContainSubstring, prefixLines("(dry-run) ", expectedSummary))
			So(memoryRedis.Keys(), ShouldResemble, allKeys)
			So(memoryRedis.TTL("app:accounts:rate-limit:OAuthTokenPerIP:127.0.0.1"), ShouldEqual, 0)
		})

		Convey("dry-run=false", func() {
			err := run(false)
			So(err, ShouldBeNil)
			So(stdout.String(), ShouldEqual, expectedOutput)
			So(stderr.String(), ShouldContainSubstring, expectedSummary)

			So(memoryRedis.Exists("app:accounts:rate-limit:SignupPerIP:127.0.0.1"), ShouldBeFalse)
			So(memoryRedis.Exists("rate-limit:MessagingSMSPerTarget:+85200000000"), ShouldBeFalse)
			So(memoryRedis.TTL("app:accounts:rate-limit:OAuthTokenPerIP:127.0.0.1"), ShouldEqual, t_future.Sub(t_now))
			So(memoryRedis.Exists("app:accounts:access-events:session-a"), ShouldBeFalse)
			So(memoryRedis.Exists("app:accounts:lockout:old"), ShouldBeFalse)
			So(memoryRedis.Exists("app:accounts:failed-attempts:purpose:old"), ShouldBeFalse)

			So(memoryRedis.Exists("app:accounts:session-list:all-dead"), ShouldBeFalse)
			mixed, err := memoryRedis.HKeys("app:accounts:session-list:mixed")
			So(err, ShouldBeNil)
			So(mixed, ShouldResemble, []string{
				"app:accounts:session:being-created",
				"app:accounts:session:live",
				"app:accounts:session:stale-expiry",
			})
			So(memoryRedis.Exists("app:accounts:session-list:unreadable"), ShouldBeTrue)

			userA, err := memoryRedis.HKeys("app:accounts:offline-grant-list:user-a")
			So(err, ShouldBeNil)
			So(userA, ShouldResemble, []string{"grant-live"})
			So(memoryRedis.Exists("app:accounts:offline-grant-list:user-b"), ShouldBeFalse)

			So(memoryRedis.Exists("app:accounts:session:rate-limit:a"), ShouldBeTrue)
			So(memoryRedis.Exists("app:accounts:session:access-events:a"), ShouldBeTrue)

			Convey("a rerun changes nothing", func() {
				keys := memoryRedis.Keys()
				err := run(false)
				So(err, ShouldBeNil)
				So(stdout.String(), ShouldEqual, "")
				So(stderr.String(), ShouldContainSubstring, "deleted=0 expiry_fixed=0 pruned=0")
				So(memoryRedis.Keys(), ShouldResemble, keys)
			})
		})

		Convey("handles the keyspace in batches", func() {
			stdout.Reset()
			stderr.Reset()
			err := Housekeep(ctx, redisClient, stdout, logger, HousekeepOptions{
				ScanCountString: "10",
				DryRun:          true,
				Now:             t_now,
			})
			So(err, ShouldBeNil)
			So(strings.Count(stderr.String(), "SCAN "), ShouldBeGreaterThan, 1)
			So(stderr.String(), ShouldContainSubstring, prefixLines("(dry-run) ", expectedSummary))
		})
	})

	Convey("housekeep with a key written during the run", t, func() {
		memoryRedis := miniredis.RunT(t)
		memoryRedis.SetTime(t_now)
		ctx := context.Background()
		redisClient := goredis.NewClient(&goredis.Options{Addr: memoryRedis.Addr()})
		otherClient := goredis.NewClient(&goredis.Options{Addr: memoryRedis.Addr()})
		stdout := &bytes.Buffer{}
		stderr := &bytes.Buffer{}
		logger := log.New(stderr, "", 0)

		_, err := redisClient.Set(ctx, "key", "0", 0).Result()
		So(err, ShouldBeNil)

		writeTimes := 0
		attempts := 0
		task := housekeepTask{
			name:  "test",
			match: func(key string) bool { return key == "key" },
			plan: func(ctx context.Context, tx *goredis.Tx, keys []string, now time.Time) ([]housekeepAction, error) {
				attempts += 1
				// Someone else writes the watched key after it is read.
				if attempts <= writeTimes {
					_, err := otherClient.Set(ctx, "key", strconv.Itoa(attempts), 0).Result()
					So(err, ShouldBeNil)
				}
				return []housekeepAction{deleteAction("key")}, nil
			},
		}
		run := func() error {
			return housekeep(ctx, redisClient, stdout, logger, HousekeepOptions{
				ScanCountString: "100",
				DryRun:          false,
				Now:             t_now,
			}, []housekeepTask{task})
		}

		Convey("retries and applies the change", func() {
			writeTimes = 1
			err := run()
			So(err, ShouldBeNil)
			So(attempts, ShouldEqual, 2)
			So(stdout.String(), ShouldEqual, "deleted key\n")
			So(memoryRedis.Exists("key"), ShouldBeFalse)
		})

		Convey("gives up and leaves the key unchanged", func() {
			writeTimes = housekeepMaxAttempts
			err := run()
			So(err, ShouldBeNil)
			So(attempts, ShouldEqual, housekeepMaxAttempts)
			So(stdout.String(), ShouldEqual, "")
			So(stderr.String(), ShouldContainSubstring, "test: keys kept changing, left for the next run: key\n")
			So(stderr.String(), ShouldContainSubstring, "done test deleted=0 expiry_fixed=0 pruned=0 skipped=0 conflict=1\n")
			v, err := memoryRedis.Get("key")
			So(err, ShouldBeNil)
			So(v, ShouldEqual, strconv.Itoa(housekeepMaxAttempts))
		})
	})
}

func prefixLines(prefix string, s string) string {
	lines := strings.SplitAfter(s, "\n")
	var b strings.Builder
	for _, line := range lines {
		if line != "" {
			b.WriteString(prefix)
			b.WriteString(line)
		}
	}
	return b.String()
}
