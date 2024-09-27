package otp

import (
	"context"
	"errors"
	"fmt"
	"time"

	goredis "github.com/go-redis/redis/v8"

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

func (s *LookupStoreRedis) Create(purpose Purpose, code string, target string, expireAt time.Time) error {
	ctx := context.Background()
	key := redisLookupKey(s.AppID, purpose, code)

	return s.Redis.WithConn(func(conn redis.Redis_6_0_Cmdable) error {
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

func (s *LookupStoreRedis) Get(purpose Purpose, code string) (target string, err error) {
	ctx := context.Background()
	key := redisLookupKey(s.AppID, purpose, code)

	err = s.Redis.WithConn(func(conn redis.Redis_6_0_Cmdable) error {
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

func (s *LookupStoreRedis) Delete(purpose Purpose, code string) error {
	ctx := context.Background()
	key := redisLookupKey(s.AppID, purpose, code)

	return s.Redis.WithConn(func(conn redis.Redis_6_0_Cmdable) error {
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
