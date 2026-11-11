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

var t_expired = time.Date(2025, 3, 5, 0, 0, 0, 0, time.UTC)
var t_now = time.Date(2025, 3, 5, 1, 0, 0, 0, time.UTC)
var t_future = time.Date(2025, 3, 5, 2, 0, 0, 0, time.UTC)

// housekeepFixture is an in-memory Redis at t_now, with helpers to write keys the way the application does.
type housekeepFixture struct {
	ctx         context.Context
	memoryRedis *miniredis.Miniredis
	redisClient *goredis.Client
	hook        *commandCountingHook
}

func newHousekeepFixture(t *testing.T) *housekeepFixture {
	memoryRedis := miniredis.RunT(t)
	memoryRedis.SetTime(t_now)
	redisClient := goredis.NewClient(&goredis.Options{Addr: memoryRedis.Addr()})
	hook := &commandCountingHook{counts: map[string]int{}}
	redisClient.AddHook(hook)
	return &housekeepFixture{
		ctx:         context.Background(),
		memoryRedis: memoryRedis,
		redisClient: redisClient,
		hook:        hook,
	}
}

func (f *housekeepFixture) set(key string, value string, ttl time.Duration) {
	_, err := f.redisClient.Set(f.ctx, key, value, ttl).Result()
	So(err, ShouldBeNil)
}

func (f *housekeepFixture) xadd(key string) {
	_, err := f.redisClient.XAdd(f.ctx, &goredis.XAddArgs{Stream: key, Values: map[string]any{"access": "{}"}}).Result()
	So(err, ShouldBeNil)
}

func (f *housekeepFixture) hsetRaw(key string, field string, value string) {
	_, err := f.redisClient.HSet(f.ctx, key, field, value).Result()
	So(err, ShouldBeNil)
}

func (f *housekeepFixture) expire(key string, ttl time.Duration) {
	ok, err := f.redisClient.Expire(f.ctx, key, ttl).Result()
	So(err, ShouldBeNil)
	So(ok, ShouldBeTrue)
}

// run runs the given tasks, or every task if none is given, and returns stdout and stderr.
func (f *housekeepFixture) run(dryRun bool, tasks ...housekeepTask) (string, string) {
	if len(tasks) == 0 {
		tasks = housekeepTasks
	}
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	f.hook.counts = map[string]int{}
	err := housekeep(f.ctx, f.redisClient, stdout, log.New(stderr, "", 0), HousekeepOptions{
		ScanCountString: "100",
		DryRun:          dryRun,
		Now:             t_now,
	}, tasks)
	So(err, ShouldBeNil)
	return stdout.String(), stderr.String()
}

func unixMilliString(t time.Time) string {
	return strconv.FormatInt(t.UnixMilli(), 10)
}

func TestHousekeep(t *testing.T) {
	Convey("Housekeep validates options", t, func() {
		f := newHousekeepFixture(t)
		logger := log.New(&bytes.Buffer{}, "", 0)

		err := Housekeep(f.ctx, f.redisClient, &bytes.Buffer{}, logger, HousekeepOptions{
			ScanCountString: "a",
			DryRun:          true,
			Now:             t_now,
		})
		So(err, ShouldBeError, "SCAN count must be an integer: a")

		err = Housekeep(f.ctx, f.redisClient, &bytes.Buffer{}, logger, HousekeepOptions{
			ScanCountString: "100",
			DryRun:          true,
			Now:             time.Time{},
		})
		So(err, ShouldBeError, "now cannot be zero")
	})

	Convey("Housekeep", t, func() {
		f := newHousekeepFixture(t)

		// One key for each task to change, and keys no task matches.
		f.set("app:accounts:rate-limit:SignupPerIP:127.0.0.1", unixMilliString(t_expired), 0)
		f.xadd("app:accounts:access-events:session")
		f.set("app:accounts:lockout:AccountAuthentication:user", "1", 0)
		f.set("app:accounts:failed-attempts:oob-otp:+85200000000", "1", 0)
		f.set("app:accounts:unrelated", "value", 0)

		expectedOutput := `deleted app:accounts:rate-limit:SignupPerIP:127.0.0.1
deleted app:accounts:access-events:session
deleted app:accounts:lockout:AccountAuthentication:user
deleted app:accounts:failed-attempts:oob-otp:+85200000000
`
		keysLeft := []string{
			"app:accounts:unrelated",
		}

		Convey("gives each key to its task and leaves the other keys alone", func() {
			stdout, stderr := f.run(false)
			So(stdout, ShouldEqual, expectedOutput)
			So(f.memoryRedis.Keys(), ShouldResemble, keysLeft)
			So(stderr, ShouldContainSubstring, "done scanned_total=5 deleted=4 expiry_fixed=0 skipped=0 conflict=0\n")
		})

		Convey("dry-run reports the changes but makes none", func() {
			keys := f.memoryRedis.Keys()
			stdout, _ := f.run(true)
			So(stdout, ShouldEqual, prefixLines("(dry-run) ", expectedOutput))
			So(f.memoryRedis.Keys(), ShouldResemble, keys)
			So(f.hook.counts["watch"], ShouldEqual, 0)
			So(f.hook.counts["multi"], ShouldEqual, 0)
		})

		Convey("only watches the keys to change, and a rerun changes nothing", func() {
			f.run(false)
			// 1 chunk for each task with changes.
			So(f.hook.counts["watch"], ShouldEqual, 4)

			stdout, stderr := f.run(false)
			So(stdout, ShouldEqual, "")
			So(stderr, ShouldContainSubstring, "deleted=0 expiry_fixed=0 skipped")
			So(f.memoryRedis.Keys(), ShouldResemble, keysLeft)
			So(f.hook.counts["watch"], ShouldEqual, 0)
			So(f.hook.counts["multi"], ShouldEqual, 0)
		})
	})

	Convey("Housekeep handles the keyspace in batches and chunks", t, func() {
		readChunkSize, writeChunkSize := housekeepReadChunkSize, housekeepWriteChunkSize
		housekeepReadChunkSize, housekeepWriteChunkSize = 4, 3
		defer func() {
			housekeepReadChunkSize, housekeepWriteChunkSize = readChunkSize, writeChunkSize
		}()

		f := newHousekeepFixture(t)
		var deleted []string
		var keysLeft []string
		for i := range 25 {
			key := "rate-limit:SignupPerIP:full-" + strconv.Itoa(100+i)
			f.set(key, unixMilliString(t_expired), 0)
			deleted = append(deleted, key)
		}
		for i := range 5 {
			key := "rate-limit:SignupPerIP:not-full-" + strconv.Itoa(100+i)
			f.set(key, unixMilliString(t_future), t_future.Sub(t_now))
			keysLeft = append(keysLeft, key)
		}

		run := func(scanCount string, dryRun bool) (string, string) {
			stdout := &bytes.Buffer{}
			stderr := &bytes.Buffer{}
			err := Housekeep(f.ctx, f.redisClient, stdout, log.New(stderr, "", 0), HousekeepOptions{
				ScanCountString: scanCount,
				DryRun:          dryRun,
				Now:             t_now,
			})
			So(err, ShouldBeNil)
			return stdout.String(), stderr.String()
		}

		// The miniredis SCAN cursor is an offset into the sorted keys, so deleting keys during a scan skips keys.
		// Real Redis does not skip keys that exist for the whole scan, so a real run over many SCAN batches
		// is tested against real Redis instead. Here SCAN batches are tested with a dry run,
		// and chunks with one SCAN batch.

		Convey("SCAN batches", func() {
			keys := f.memoryRedis.Keys()
			stdout, stderr := run("10", true)
			So(strings.Count(stderr, "SCAN "), ShouldBeGreaterThan, 1)
			So(stdout, ShouldEqual, "(dry-run) deleted "+strings.Join(deleted, "\n(dry-run) deleted ")+"\n")
			So(f.memoryRedis.Keys(), ShouldResemble, keys)
		})

		Convey("read and write chunks", func() {
			stdout, stderr := run("100", false)
			So(strings.Count(stderr, "SCAN "), ShouldEqual, 1)
			So(stdout, ShouldEqual, "deleted "+strings.Join(deleted, "\ndeleted ")+"\n")
			So(f.memoryRedis.Keys(), ShouldResemble, keysLeft)
			// 30 keys in read chunks of 4: 6 full chunks of 4 to delete, and 1 chunk of 1.
			// They are written in write chunks of 3: 2 for each full chunk, and 1 for the last.
			So(f.hook.counts["watch"], ShouldEqual, 13)
		})
	})

	Convey("Housekeep with a key written during the run", t, func() {
		f := newHousekeepFixture(t)
		otherClient := goredis.NewClient(&goredis.Options{Addr: f.memoryRedis.Addr()})
		f.set("key", "0", 0)

		writeTimes := 0
		attempts := 0
		fixedBeforeWatch := false
		task := housekeepTask{
			name:  "test",
			match: func(key string) bool { return key == "key" },
			plan: func(ctx context.Context, r housekeepReader, keys []string, now time.Time) ([]housekeepAction, error) {
				if _, watched := r.(*goredis.Tx); !watched {
					if fixedBeforeWatch {
						// Someone else fixes the key between the two reads.
						_, err := otherClient.Set(ctx, "key", "fixed", 0).Result()
						So(err, ShouldBeNil)
					}
					return []housekeepAction{deleteAction("key")}, nil
				}

				attempts += 1
				v, err := otherClient.Get(ctx, "key").Result()
				So(err, ShouldBeNil)
				if v == "fixed" {
					return nil, nil
				}
				// Someone else writes the watched key after it is read.
				if attempts <= writeTimes {
					_, err := otherClient.Set(ctx, "key", strconv.Itoa(attempts), 0).Result()
					So(err, ShouldBeNil)
				}
				return []housekeepAction{deleteAction("key")}, nil
			},
		}

		Convey("retries and applies the change", func() {
			writeTimes = 1
			stdout, _ := f.run(false, task)
			So(attempts, ShouldEqual, 2)
			So(stdout, ShouldEqual, "deleted key\n")
			So(f.memoryRedis.Exists("key"), ShouldBeFalse)
		})

		Convey("gives up and leaves the key unchanged", func() {
			writeTimes = housekeepMaxAttempts
			stdout, stderr := f.run(false, task)
			So(attempts, ShouldEqual, housekeepMaxAttempts)
			So(stdout, ShouldEqual, "")
			So(stderr, ShouldContainSubstring, "test: keys kept changing, left for the next run: key\n")
			So(stderr, ShouldContainSubstring, "done test deleted=0 expiry_fixed=0 skipped=0 conflict=1\n")
			v, err := f.memoryRedis.Get("key")
			So(err, ShouldBeNil)
			So(v, ShouldEqual, strconv.Itoa(housekeepMaxAttempts))
		})

		Convey("skips a key that needs no change when read again", func() {
			fixedBeforeWatch = true
			stdout, stderr := f.run(false, task)
			So(attempts, ShouldEqual, 1)
			So(stdout, ShouldEqual, "")
			So(stderr, ShouldContainSubstring, "done test deleted=0 expiry_fixed=0 skipped=1 conflict=0\n")
			v, err := f.memoryRedis.Get("key")
			So(err, ShouldBeNil)
			So(v, ShouldEqual, "fixed")
		})
	})
}

type commandCountingHook struct {
	counts map[string]int
}

func (h *commandCountingHook) DialHook(next goredis.DialHook) goredis.DialHook {
	return next
}

func (h *commandCountingHook) ProcessHook(next goredis.ProcessHook) goredis.ProcessHook {
	return func(ctx context.Context, cmd goredis.Cmder) error {
		h.counts[cmd.Name()] += 1
		return next(ctx, cmd)
	}
}

func (h *commandCountingHook) ProcessPipelineHook(next goredis.ProcessPipelineHook) goredis.ProcessPipelineHook {
	return func(ctx context.Context, cmds []goredis.Cmder) error {
		for _, cmd := range cmds {
			h.counts[cmd.Name()] += 1
		}
		return next(ctx, cmds)
	}
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
