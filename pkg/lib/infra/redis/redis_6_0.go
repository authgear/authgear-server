package redis

import (
	"context"
	"time"

	goredis "github.com/go-redis/redis/v8"
)

type Redis_6_0_Cmdable interface {
	Del(ctx context.Context, keys ...string) *goredis.IntCmd

	Get(ctx context.Context, key string) *goredis.StringCmd

	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *goredis.StatusCmd
	SetEX(ctx context.Context, key string, value interface{}, expiration time.Duration) *goredis.StatusCmd
	SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) *goredis.BoolCmd
	SetXX(ctx context.Context, key string, value interface{}, expiration time.Duration) *goredis.BoolCmd

	Expire(ctx context.Context, key string, expiration time.Duration) *goredis.BoolCmd
	ExpireAt(ctx context.Context, key string, tm time.Time) *goredis.BoolCmd

	Incr(ctx context.Context, key string) *goredis.IntCmd
	IncrBy(ctx context.Context, key string, value int64) *goredis.IntCmd

	PExpireAt(ctx context.Context, key string, tm time.Time) *goredis.BoolCmd

	XAdd(ctx context.Context, a *goredis.XAddArgs) *goredis.StringCmd

	HDel(ctx context.Context, key string, fields ...string) *goredis.IntCmd
	HSet(ctx context.Context, key string, values ...interface{}) *goredis.IntCmd
	HGetAll(ctx context.Context, key string) *goredis.StringStringMapCmd

	LPush(ctx context.Context, key string, values ...interface{}) *goredis.IntCmd
	BRPop(ctx context.Context, timeout time.Duration, keys ...string) *goredis.StringSliceCmd

	// HyperLogLog.
	PFCount(ctx context.Context, keys ...string) *goredis.IntCmd
	PFAdd(ctx context.Context, key string, els ...interface{}) *goredis.IntCmd

	// For lua script
	Eval(ctx context.Context, script string, keys []string, args ...interface{}) *goredis.Cmd
	EvalSha(ctx context.Context, sha1 string, keys []string, args ...interface{}) *goredis.Cmd
	ScriptExists(ctx context.Context, hashes ...string) *goredis.BoolSliceCmd
	ScriptLoad(ctx context.Context, script string) *goredis.StringCmd
}
