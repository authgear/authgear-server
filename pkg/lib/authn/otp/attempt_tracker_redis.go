package otp

import (
	"context"
	"errors"
	"fmt"

	goredis "github.com/go-redis/redis/v8"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/redis/appredis"
	"github.com/authgear/authgear-server/pkg/util/clock"
)

type AttemptTrackerRedis struct {
	Redis *appredis.Handle
	AppID config.AppID
	Clock clock.Clock
}

func (s *AttemptTrackerRedis) ResetFailedAttempts(purpose string, target string) error {
	ctx := context.Background()

	return s.Redis.WithConn(func(conn *goredis.Conn) error {
		_, err := conn.Del(ctx, redisFailedAttemptsKey(s.AppID, purpose, target)).Result()
		return err
	})
}

func (s *AttemptTrackerRedis) GetFailedAttempts(purpose string, target string) (int, error) {
	ctx := context.Background()

	var failedAttempts int
	err := s.Redis.WithConn(func(conn *goredis.Conn) (err error) {
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

func (s *AttemptTrackerRedis) IncrementFailedAttempts(purpose string, target string) (int, error) {
	ctx := context.Background()

	var failedAttempts int64
	err := s.Redis.WithConn(func(conn *goredis.Conn) (err error) {
		failedAttempts, err = conn.Incr(ctx, redisFailedAttemptsKey(s.AppID, purpose, target)).Result()
		if err != nil {
			return err
		}

		return nil
	})
	return int(failedAttempts), err
}

func redisFailedAttemptsKey(appID config.AppID, purpose string, target string) string {
	return fmt.Sprintf("app:%s:failed-attempts:%s:%s", appID, purpose, target)
}
