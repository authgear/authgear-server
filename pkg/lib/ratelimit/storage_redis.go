package ratelimit

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	redigo "github.com/gomodule/redigo/redis"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/redis"
)

type StorageRedis struct {
	AppID config.AppID
	Redis *redis.Handle
}

func (s *StorageRedis) WithConn(fn func(StorageConn) error) error {
	return s.Redis.WithConn(func(conn redis.Conn) error {
		return fn(storageRedisConn{AppID: s.AppID, Conn: conn})
	})
}

type storageRedisConn struct {
	AppID config.AppID
	Conn  redigo.Conn
}

func (s storageRedisConn) TakeToken(bucket Bucket, now time.Time) (int, error) {
	tokenTaken, err := redigo.Int(s.Conn.Do("HINCRBY", redisBucketKey(s.AppID, bucket), "token_taken", 1))
	if err != nil {
		return 0, err
	}
	tokens := bucket.Size - tokenTaken

	// Populate reset time if not yet exists.
	resetTime := now.Add(bucket.ResetPeriod).UnixNano()
	created, err := redigo.Bool(s.Conn.Do("HSETNX", redisBucketKey(s.AppID, bucket), "reset_time", resetTime))
	if err != nil {
		return 0, err
	}
	if created {
		// Ignore error
		_, _ = s.Conn.Do("PEXPIREAT", redisBucketKey(s.AppID, bucket), resetTime/1000000)
	}

	return tokens, nil
}

func (s storageRedisConn) GetResetTime(bucket Bucket, now time.Time) (time.Time, error) {
	resetTime, err := redigo.String(s.Conn.Do("HGET", redisBucketKey(s.AppID, bucket), "reset_time"))
	if errors.Is(err, redigo.ErrNil) {
		// Reset time is not present, default to now to avoid stuck state
		return now, nil
	} else if err != nil {
		return time.Time{}, err
	}

	nano, err := strconv.ParseInt(resetTime, 10, 64)
	if err != nil {
		// Invalid reset time, default to now to avoid stuck state
		return now, nil
	}
	return time.Unix(0, nano), nil
}

func (s storageRedisConn) Reset(bucket Bucket, now time.Time) error {
	// Delete bucket data, so TakeToken would regenerate the bucket.
	_, _ = s.Conn.Do("DEL", redisBucketKey(s.AppID, bucket))
	return nil
}

func redisBucketKey(appID config.AppID, bucket Bucket) string {
	return fmt.Sprintf("app:%s:rate-limit:%s", appID, bucket.Key)
}
