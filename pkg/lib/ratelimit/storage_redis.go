package ratelimit

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	goredis "github.com/go-redis/redis/v8"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/redis/appredis"
)

type StorageRedis struct {
	AppID config.AppID
	Redis *appredis.Handle
}

func (s *StorageRedis) WithConn(fn func(StorageConn) error) error {
	return s.Redis.WithConn(func(conn *goredis.Conn) error {
		return fn(storageRedisConn{AppID: s.AppID, Conn: conn})
	})
}

type storageRedisConn struct {
	AppID config.AppID
	Conn  *goredis.Conn
}

func (s storageRedisConn) TakeToken(spec BucketSpec, now time.Time, delta int) (int, error) {
	// FIXME: token may underflow to negative?

	ctx := context.Background()
	key := redisBucketKey(s.AppID, spec)

	tokenTaken, err := s.Conn.HIncrBy(ctx, key, "token_taken", int64(delta)).Result()
	if err != nil {
		return 0, err
	}

	tokens := int64(spec.Burst) - tokenTaken

	// Populate reset time if not yet exists.
	resetTime := now
	if tokenTaken > 0 {
		resetTime = resetTime.Add(spec.Period)
	}

	created, err := s.Conn.HSetNX(ctx, key, "reset_time", resetTime.UnixNano()).Result()
	if err != nil {
		return 0, err
	}

	if created {
		// Ignore error
		_, _ = s.Conn.PExpireAt(ctx, key, resetTime).Result()
	}

	return int(tokens), nil
}

func (s storageRedisConn) GetResetTime(spec BucketSpec, now time.Time) (time.Time, error) {
	ctx := context.Background()
	resetTime, err := s.Conn.HGet(ctx, redisBucketKey(s.AppID, spec), "reset_time").Result()
	if errors.Is(err, goredis.Nil) {
		return now, nil
	} else if err != nil {
		return time.Time{}, err
	}

	nano, err := strconv.ParseInt(resetTime, 10, 64)
	if err != nil {
		// Invalid reset time, default to now to avoid stuck state
		return now, nil
	}
	return time.Unix(0, nano).UTC(), nil
}

func (s storageRedisConn) Reset(bucket BucketSpec, now time.Time) error {
	// Delete bucket data, so TakeToken would regenerate the bucket.
	ctx := context.Background()
	_, _ = s.Conn.Del(ctx, redisBucketKey(s.AppID, bucket)).Result()
	return nil
}

func redisBucketKey(appID config.AppID, spec BucketSpec) string {
	if spec.IsGlobal {
		return fmt.Sprintf("rate-limit:%s", spec.Key())
	}
	return fmt.Sprintf("app:%s:rate-limit:%s", appID, spec.Key())
}
