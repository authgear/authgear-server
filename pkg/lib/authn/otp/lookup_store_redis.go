package otp

import (
	"context"
	"errors"
	"fmt"
	"time"

	goredis "github.com/redis/go-redis/v9"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/redis"
	"github.com/authgear/authgear-server/pkg/lib/infra/redis/appredis"
	"github.com/authgear/authgear-server/pkg/util/clock"
)

type LookupStoreRedis struct {
	Redis *appredis.Handle
	AppID config.AppID
	Clock clock.Clock
}

var _ LookupStore = &LookupStoreRedis{}

func (s *LookupStoreRedis) Create(ctx context.Context, purpose Purpose, code string, target string, expireAt time.Time) error {
	key := redisLookupKey(s.AppID, purpose, code)

	return s.Redis.WithConnContext(ctx, func(ctx context.Context, conn redis.Redis_6_0_Cmdable) error {
		ttl := expireAt.Sub(s.Clock.NowUTC())

		_, err := conn.SetNX(ctx, key, target, ttl).Result()
		if errors.Is(err, goredis.Nil) {
			return errors.New("duplicated code")
		} else if err != nil {
			return err
		}

		return nil
	})
}

func (s *LookupStoreRedis) Get(ctx context.Context, purpose Purpose, code string) (target string, err error) {
	key := redisLookupKey(s.AppID, purpose, code)

	err = s.Redis.WithConnContext(ctx, func(ctx context.Context, conn redis.Redis_6_0_Cmdable) error {
		target, err = conn.Get(ctx, key).Result()
		if errors.Is(err, goredis.Nil) {
			return ErrCodeNotFound
		} else if err != nil {
			return err
		}

		return nil
	})
	return
}

func (s *LookupStoreRedis) Delete(ctx context.Context, purpose Purpose, code string) error {
	key := redisLookupKey(s.AppID, purpose, code)

	return s.Redis.WithConnContext(ctx, func(ctx context.Context, conn redis.Redis_6_0_Cmdable) error {
		_, err := conn.Del(ctx, key).Result()
		if err != nil {
			return err
		}
		return err
	})
}

func redisLookupKey(appID config.AppID, purpose Purpose, code string) string {
	return fmt.Sprintf("app:%s:otp-lookup:%s:%s", appID, purpose, code)
}
