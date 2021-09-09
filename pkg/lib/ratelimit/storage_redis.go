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

func (s storageRedisConn) TakeToken(bucket Bucket, now time.Time) (int, error) {
	ctx := context.Background()
	tokenTaken, err := s.Conn.HIncrBy(ctx, redisBucketKey(s.AppID, bucket), "token_taken", 1).Result()
	if err != nil {
		return 0, err
	}

	tokens := int64(bucket.Size) - tokenTaken

	// Populate reset time if not yet exists.
	resetTime := now.Add(bucket.ResetPeriod)

	created, err := s.Conn.HSetNX(ctx, redisBucketKey(s.AppID, bucket), "reset_time", resetTime.UnixNano()).Result()
	if err != nil {
		return 0, err
	}

	if created {
		// Ignore error
		_, _ = s.Conn.PExpireAt(ctx, redisBucketKey(s.AppID, bucket), resetTime).Result()
	}

	return int(tokens), nil
}

func (s storageRedisConn) CheckToken(bucket Bucket) (int, error) {
	ctx := context.Background()
	tokenTaken, err := s.Conn.HGet(ctx, redisBucketKey(s.AppID, bucket), "token_taken").Int64()
	if errors.Is(err, goredis.Nil) {
		tokenTaken = 0
	} else if err != nil {
		return 0, err
	}

	tokens := int64(bucket.Size) - tokenTaken
	return int(tokens), nil
}

func (s storageRedisConn) GetResetTime(bucket Bucket, now time.Time) (time.Time, error) {
	ctx := context.Background()
	resetTime, err := s.Conn.HGet(ctx, redisBucketKey(s.AppID, bucket), "reset_time").Result()
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

func (s storageRedisConn) Reset(bucket Bucket, now time.Time) error {
	// Delete bucket data, so TakeToken would regenerate the bucket.
	ctx := context.Background()
	_, _ = s.Conn.Del(ctx, redisBucketKey(s.AppID, bucket)).Result()
	return nil
}

func redisBucketKey(appID config.AppID, bucket Bucket) string {
	return fmt.Sprintf("app:%s:rate-limit:%s", appID, bucket.Key)
}
