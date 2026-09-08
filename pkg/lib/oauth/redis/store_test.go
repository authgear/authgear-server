package redis_test

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/redis"
	"github.com/authgear/authgear-server/pkg/lib/infra/redis/appredis"
	"github.com/authgear/authgear-server/pkg/lib/oauth"
	oauthredis "github.com/authgear/authgear-server/pkg/lib/oauth/redis"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/duration"

	. "github.com/smartystreets/goconvey/convey"
)

func TestConsumeGrant(t *testing.T) {
	Convey("consuming a single use grant", t, func() {
		mr := miniredis.RunT(t)

		connectionOptions := redis.ConnectionOptions{
			RedisURL:          "redis://" + mr.Addr(),
			MaxOpenConnection: func(i int) *int { return &i }(10),
			MaxIdleConnection: func(i int) *int { return &i }(5),
			IdleConnectionTimeout: func(d time.Duration) *config.DurationSeconds {
				ds := config.DurationSeconds(d.Seconds())
				return &ds
			}(duration.UserInteraction),
			MaxConnectionLifetime: func(d time.Duration) *config.DurationSeconds {
				ds := config.DurationSeconds(d.Seconds())
				return &ds
			}(duration.UserInteraction),
		}

		appID := "app-id"
		clk := clock.NewMockClock()
		store := &oauthredis.Store{
			Redis: &appredis.Handle{Handle: redis.NewHandle(redis.NewPool(), connectionOptions)},
			AppID: config.AppID(appID),
			Clock: clk,
		}

		ctx := context.Background()

		Convey("the code grant is consumable exactly once", func() {
			grant := &oauth.CodeGrant{
				AppID:    appID,
				CodeHash: "code-hash",
				ExpireAt: clk.NowUTC().Add(time.Hour),
			}
			So(store.CreateCodeGrant(ctx, grant), ShouldBeNil)

			// DEL removed the key, so this call is the one that spent it.
			So(store.ConsumeCodeGrant(ctx, grant), ShouldBeNil)

			// A second caller -- the loser of a race, or a replay -- removed
			// nothing, and must be told so rather than being allowed to
			// proceed as if it had spent the code.
			So(store.ConsumeCodeGrant(ctx, grant), ShouldBeError, oauth.ErrGrantNotFound)

			_, err := store.GetCodeGrant(ctx, "code-hash")
			So(err, ShouldBeError, oauth.ErrGrantNotFound)
		})

		Convey("deleting a code grant that was never created reports not found", func() {
			So(store.ConsumeCodeGrant(ctx, &oauth.CodeGrant{
				AppID:    appID,
				CodeHash: "never-created",
			}), ShouldBeError, oauth.ErrGrantNotFound)
		})

		Convey("the settings action grant is consumable exactly once", func() {
			grant := &oauth.SettingsActionGrant{
				AppID:    appID,
				CodeHash: "settings-code-hash",
				ExpireAt: clk.NowUTC().Add(time.Hour),
			}
			So(store.CreateSettingsActionGrant(ctx, grant), ShouldBeNil)

			So(store.ConsumeSettingsActionGrant(ctx, grant), ShouldBeNil)
			So(store.ConsumeSettingsActionGrant(ctx, grant), ShouldBeError, oauth.ErrGrantNotFound)
		})
	})
}
