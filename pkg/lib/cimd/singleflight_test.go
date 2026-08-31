package cimd_test

import (
	"context"
	"testing"

	"github.com/alicebob/miniredis/v2"
	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/lib/cimd"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/redis"
	"github.com/authgear/authgear-server/pkg/lib/infra/redis/appredis"
)

func newTestFetchSingleFlight(t *testing.T) (*cimd.FetchSingleFlight, *miniredis.Miniredis) {
	mr := miniredis.RunT(t)
	pool := redis.NewPool()
	rh := redis.NewHandle(pool, redis.ConnectionOptions{
		RedisURL:              "redis://" + mr.Addr(),
		MaxOpenConnection:     func(i int) *int { return &i }(10),
		MaxIdleConnection:     func(i int) *int { return &i }(5),
		IdleConnectionTimeout: func(d config.DurationSeconds) *config.DurationSeconds { return &d }(300),
		MaxConnectionLifetime: func(d config.DurationSeconds) *config.DurationSeconds { return &d }(900),
	})

	return &cimd.FetchSingleFlight{
		Redis: &appredis.Handle{Handle: rh},
		AppID: "test-app",
	}, mr
}

func TestFetchSingleFlightAcquire(t *testing.T) {
	ctx := context.Background()

	Convey("FetchSingleFlight.Acquire", t, func() {
		sf, mr := newTestFetchSingleFlight(t)

		Convey("the first caller for a client_id acquires", func() {
			acquired, err := sf.Acquire(ctx, "cimd-fetch", "https://mcp-client.example.com/oauth/client-metadata.json")
			So(err, ShouldBeNil)
			So(acquired, ShouldBeTrue)
		})

		Convey("a second concurrent caller for the same client_id does not acquire", func() {
			clientID := "https://mcp-client.example.com/oauth/client-metadata.json"
			first, err := sf.Acquire(ctx, "cimd-fetch", clientID)
			So(err, ShouldBeNil)
			So(first, ShouldBeTrue)

			second, err := sf.Acquire(ctx, "cimd-fetch", clientID)
			So(err, ShouldBeNil)
			So(second, ShouldBeFalse)
		})

		Convey("two different client_ids do not contend", func() {
			a, err := sf.Acquire(ctx, "cimd-fetch", "https://a.example.com/x")
			So(err, ShouldBeNil)
			So(a, ShouldBeTrue)

			b, err := sf.Acquire(ctx, "cimd-fetch", "https://b.example.com/x")
			So(err, ShouldBeNil)
			So(b, ShouldBeTrue)
		})

		Convey("the lock expires, letting a later caller acquire again", func() {
			clientID := "https://mcp-client.example.com/oauth/client-metadata.json"
			first, err := sf.Acquire(ctx, "cimd-fetch", clientID)
			So(err, ShouldBeNil)
			So(first, ShouldBeTrue)

			mr.FastForward(11 * 1000000000) // 11s > fetchLockTTL (10s)

			again, err := sf.Acquire(ctx, "cimd-fetch", clientID)
			So(err, ShouldBeNil)
			So(again, ShouldBeTrue)
		})

		Convey("the key is hashed, not the raw client_id: a colon in the URL cannot collide with the key namespace", func() {
			// If clientID were interpolated raw, a client_id containing ":"
			// could forge a key that collides with a different app_id or a
			// different Redis key prefix. Assert indirectly: the stored key
			// name never contains the raw client_id substring.
			clientID := "https://evil.example.com/a:app:other-app:cimd-fetch:x"
			_, err := sf.Acquire(ctx, "cimd-fetch", clientID)
			So(err, ShouldBeNil)

			for _, k := range mr.Keys() {
				So(k, ShouldNotContainSubstring, clientID)
			}
		})

		Convey("the same client_id under two different purposes does not contend -- the document fetch and the logo fetch locks are independent", func() {
			clientID := "https://mcp-client.example.com/oauth/client-metadata.json"
			docLock, err := sf.Acquire(ctx, "cimd-fetch", clientID)
			So(err, ShouldBeNil)
			So(docLock, ShouldBeTrue)

			logoLock, err := sf.Acquire(ctx, "cimd-logo-fetch", clientID)
			So(err, ShouldBeNil)
			So(logoLock, ShouldBeTrue)
		})
	})
}
