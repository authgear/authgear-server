package ratelimit

import (
	"errors"
	"strconv"
	"time"

	redigo "github.com/gomodule/redigo/redis"

	"github.com/authgear/authgear-server/pkg/lib/infra/redis"
)

type StorageRedis struct {
	Redis *redis.Handle
}

func (s *StorageRedis) WithConn(fn func(StorageConn) error) error {
	return s.Redis.WithConn(func(conn redis.Conn) error {
		return fn(storageRedisConn{Conn: conn})
	})
}

type storageRedisConn struct {
	Conn redigo.Conn
}

func (s storageRedisConn) TakeToken(bucket Bucket, now time.Time) (int, error) {
	tokenTaken, err := redigo.Int(s.Conn.Do("HINCRBY", redisBucketKey(bucket), "token_taken", 1))
	if err != nil {
		return 0, err
	}
	tokens := bucket.Size - tokenTaken

	// Populate reset time if not yet exists.
	resetTime := now.Add(bucket.ResetPeriod).UnixNano()
	created, err := redigo.Bool(s.Conn.Do("HSETNX", redisBucketKey(bucket), "reset_time", resetTime))
	if err != nil {
		return 0, err
	}
	if created {
		// Ignore error
		_, _ = s.Conn.Do("PEXPIREAT", redisBucketKey(bucket), resetTime/1000000)
	}

	return tokens, nil
}

func (s storageRedisConn) GetResetTime(bucket Bucket, now time.Time) (time.Time, error) {
	resetTime, err := redigo.String(s.Conn.Do("HGET", redisBucketKey(bucket), "reset_time"))
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
	resetTime := now.Add(bucket.ResetPeriod).UnixNano()
	_, err := s.Conn.Do("HSET", redisBucketKey(bucket), "token_taken", 0, "reset_time", resetTime)
	if err != nil {
		return err
	}
	// Ignore error
	_, _ = s.Conn.Do("PEXPIREAT", redisBucketKey(bucket), resetTime/1000000)
	return nil
}

func redisBucketKey(bucket Bucket) string {
	return "rate-limit:" + bucket.Key
}
