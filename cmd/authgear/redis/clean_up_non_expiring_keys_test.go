package redis

import (
	"bytes"
	"context"
	"log"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	goredis "github.com/go-redis/redis/v8"
	. "github.com/smartystreets/goconvey/convey"
)

func testCleanUpNonExpiringKeysSetupFixture(ctx context.Context, redisClient *goredis.Client) {
	var err error

	_, err = redisClient.Set(ctx, "app:accounts:access-events:idpsession-a", "", 0).Result()
	So(err, ShouldBeNil)

	_, err = redisClient.Set(ctx, "app:accounts:access-events:idpsession-b", "", 0).Result()
	So(err, ShouldBeNil)

	_, err = redisClient.Set(ctx, "app:accounts:access-events:offline-grant-a", "", 0).Result()
	So(err, ShouldBeNil)

	_, err = redisClient.Set(ctx, "app:accounts:access-events:offline-grant-b", "", 0).Result()
	So(err, ShouldBeNil)

	_, err = redisClient.Set(ctx, "app:accounts:session:idpsession-a", "", 5*time.Second).Result()
	So(err, ShouldBeNil)

	_, err = redisClient.Set(ctx, "app:accounts:offline-grant:offline-grant-a", "", 5*time.Second).Result()
	So(err, ShouldBeNil)
}

func TestCleanUpNonExpiringKeysDryRunTrue(t *testing.T) {
	memoryRedis := miniredis.RunT(t)

	Convey("CleanUpNonExpiringKeys dry-run=true", t, func() {
		ctx := context.Background()
		redisClient := goredis.NewClient(&goredis.Options{Addr: memoryRedis.Addr()})

		testCleanUpNonExpiringKeysSetupFixture(ctx, redisClient)

		dryRun := true
		stdout := &bytes.Buffer{}
		stderr := &bytes.Buffer{}
		logger := log.New(stderr, "", 0)
		err := CleanUpNonExpiringKeys(ctx, redisClient, dryRun, stdout, logger)
		So(err, ShouldBeNil)
		So(stderr.String(), ShouldEqual, `SCAN with cursor 0: 4
done SCAN: 4
would delete app:accounts:access-events:idpsession-b
would delete app:accounts:access-events:offline-grant-b
`)
		So(stdout.String(), ShouldEqual, `app:accounts:access-events:idpsession-b
app:accounts:access-events:offline-grant-b
`)
	})
}

func TestCleanUpNonExpiringKeysDryRunFalse(t *testing.T) {
	memoryRedis := miniredis.RunT(t)

	Convey("CleanUpNonExpiringKeys dry-run=false", t, func() {
		ctx := context.Background()
		redisClient := goredis.NewClient(&goredis.Options{Addr: memoryRedis.Addr()})

		testCleanUpNonExpiringKeysSetupFixture(ctx, redisClient)

		dryRun := false
		stdout := &bytes.Buffer{}
		stderr := &bytes.Buffer{}
		logger := log.New(stderr, "", 0)
		err := CleanUpNonExpiringKeys(ctx, redisClient, dryRun, stdout, logger)
		So(err, ShouldBeNil)
		So(stderr.String(), ShouldEqual, `SCAN with cursor 0: 4
done SCAN: 4
deleted app:accounts:access-events:idpsession-b
deleted app:accounts:access-events:offline-grant-b
`)
		So(stdout.String(), ShouldEqual, `app:accounts:access-events:idpsession-b
app:accounts:access-events:offline-grant-b
`)
	})
}
