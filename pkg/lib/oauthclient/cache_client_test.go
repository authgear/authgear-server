package oauthclient_test

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/redis"
	"github.com/authgear/authgear-server/pkg/lib/infra/redis/appredis"
	"github.com/authgear/authgear-server/pkg/lib/oauthclient"
	"github.com/authgear/authgear-server/pkg/util/clock"
)

func newTestClientCache(t *testing.T) (*oauthclient.ClientCache, *miniredis.Miniredis) {
	mr := miniredis.RunT(t)
	pool := redis.NewPool()
	rh := redis.NewHandle(pool, redis.ConnectionOptions{
		RedisURL:              "redis://" + mr.Addr(),
		MaxOpenConnection:     func(i int) *int { return &i }(10),
		MaxIdleConnection:     func(i int) *int { return &i }(5),
		IdleConnectionTimeout: func(d config.DurationSeconds) *config.DurationSeconds { return &d }(300),
		MaxConnectionLifetime: func(d config.DurationSeconds) *config.DurationSeconds { return &d }(900),
	})

	return &oauthclient.ClientCache{
		Redis: &appredis.Handle{Handle: rh},
		AppID: "test-app",
		Clock: clock.NewMockClockAt("2026-08-17T00:00:00Z"),
	}, mr
}

func TestClientCache(t *testing.T) {
	ctx := context.Background()

	Convey("ClientCache", t, func() {
		cache, mr := newTestClientCache(t)
		clientName := "PR #123 preview"
		client := &oauthclient.Client{
			ID:              "row-id",
			ClientID:        "dcrc_test",
			Source:          model.OAuthClientSourceDCR,
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
			Kind:            model.OAuthClientKindThirdParty,
			ApplicationType: "web",
			ClientName:      &clientName,
			RedirectURIs:    []string{"https://example.com/callback"},
			GrantTypes:      []string{"authorization_code", "refresh_token"},
			ResponseTypes:   []string{"code"},
		}

		Convey("not cached at all returns found=false", func() {
			c, found, err := cache.Get(ctx, "dcrc_unknown")
			So(err, ShouldBeNil)
			So(found, ShouldBeFalse)
			So(c, ShouldBeNil)
		})

		Convey("Set then Get round-trips the client", func() {
			err := cache.Set(ctx, client)
			So(err, ShouldBeNil)

			c, found, err := cache.Get(ctx, client.ClientID)
			So(err, ShouldBeNil)
			So(found, ShouldBeTrue)
			So(c, ShouldNotBeNil)
			So(c.ClientID, ShouldEqual, client.ClientID)
			So(*c.ClientName, ShouldEqual, clientName)
		})

		Convey("SetNotFound then Get distinguishes a cached negative from not-cached", func() {
			err := cache.SetNotFound(ctx, "dcrc_missing")
			So(err, ShouldBeNil)

			c, found, err := cache.Get(ctx, "dcrc_missing")
			So(err, ShouldBeNil)
			So(found, ShouldBeTrue)
			So(c, ShouldBeNil)
		})

		Convey("Delete removes a cached entry", func() {
			err := cache.Set(ctx, client)
			So(err, ShouldBeNil)

			err = cache.Delete(ctx, client.ClientID)
			So(err, ShouldBeNil)

			c, found, err := cache.Get(ctx, client.ClientID)
			So(err, ShouldBeNil)
			So(found, ShouldBeFalse)
			So(c, ShouldBeNil)
		})

		Convey("the marshalled payload is the Client row, not a resolved config", func() {
			err := cache.Set(ctx, client)
			So(err, ShouldBeNil)

			// Inspect the raw bytes actually persisted in Redis, not a
			// round-trip through Get() — oauthclient.Client has no
			// token-lifetime fields at all, so this is a genuine assertion
			// about what is stored, not merely about the Go struct's shape.
			keys := mr.Keys()
			So(keys, ShouldHaveLength, 1)
			raw, err := mr.Get(keys[0])
			So(err, ShouldBeNil)
			So(raw, ShouldContainSubstring, `"ClientID":"dcrc_test"`)
			So(raw, ShouldNotContainSubstring, "AccessTokenLifetime")
			So(raw, ShouldNotContainSubstring, "RefreshTokenLifetime")
		})
	})
}
