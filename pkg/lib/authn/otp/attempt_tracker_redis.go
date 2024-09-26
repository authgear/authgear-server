package otp

import (
	"context"
	"errors"
	"fmt"

	goredis "github.com/go-redis/redis/v8"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/redis"
	"github.com/authgear/authgear-server/pkg/lib/infra/redis/appredis"
	"github.com/authgear/authgear-server/pkg/util/clock"
)

type AttemptTrackerRedis struct {
	Redis *appredis.Handle
	AppID config.AppID
	Clock clock.Clock
}

func (s *AttemptTrackerRedis) ResetFailedAttempts(kind Kind, target string) error {
	ctx := context.Background()
	purpose := kind.Purpose()

	return s.Redis.WithConn(func(conn redis.Redis_6_0_Cmdable) error {
		_, err := conn.Del(ctx, redisFailedAttemptsKey(s.AppID, purpose, target)).Result()
		return err
	})
}

func (s *AttemptTrackerRedis) GetFailedAttempts(kind Kind, target string) (int, error) {
	ctx := context.Background()
	purpose := kind.Purpose()

	var failedAttempts int
	err := s.Redis.WithConn(func(conn redis.Redis_6_0_Cmdable) (err error) {
		failedAttempts, err = conn.Get(ctx, redisFailedAttemptsKey(s.AppID, purpose, target)).Int()
		if errors.Is(err, goredis.Nil) {
			failedAttempts = 0
			return nil
		} else if err != nil {
			return err
		}

		return nil
	})
	return failedAttempts, err
}

func (s *AttemptTrackerRedis) IncrementFailedAttempts(kind Kind, target string) (int, error) {
	ctx := context.Background()

	purpose := kind.Purpose()
	key := redisFailedAttemptsKey(s.AppID, purpose, target)
	// Whenever we increment the number of failed attempts,
	// we extend the expiration to be the valid period of the OTP.
	// This ensures the number of failed attempts outlives the OTP.
	expiration := kind.ValidPeriod()

	var failedAttempts int64
	err := s.Redis.WithConn(func(conn redis.Redis_6_0_Cmdable) (err error) {
		failedAttempts, err = conn.Incr(ctx, key).Result()
		if err != nil {
			return err
		}

		_, err = conn.Expire(ctx, key, expiration).Result()
		if err != nil {
			return err
		}

		return nil
	})
	return int(failedAttempts), err
}

func redisFailedAttemptsKey(appID config.AppID, purpose Purpose, target string) string {
	return fmt.Sprintf("app:%s:failed-attempts:%s:%s", appID, purpose, target)
}
