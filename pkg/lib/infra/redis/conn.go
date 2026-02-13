package redis

import (
	"context"
	"fmt"
	"time"

	goredis "github.com/redis/go-redis/v9"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	semconv "go.opentelemetry.io/otel/semconv/v1.34.0"
	"go.opentelemetry.io/otel/trace"
)

var tracer = otel.Tracer("github.com/authgear/authgear-server/pkg/lib/infra/redis")

var _ Redis_6_0_Cmdable = (*otelRedisConn)(nil)

// This struct wraps a *goredis.Conn, ensures each command created an otel span for tracing
type otelRedisConn struct {
	conn *goredis.Conn
}

func (c *otelRedisConn) withSpan(ctx context.Context, operation string, fn func(ctx context.Context) error) {
	ctx, span := tracer.Start(ctx, "Redis "+operation, trace.WithAttributes(
		attribute.String("db.system", "redis"),
		semconv.DBOperationName(operation),
	))
	var err error
	defer func() {
		if r := recover(); r != nil {
			err := fmt.Errorf("panic: %v", r)
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			span.End()
			panic(r)
		}
		if err != nil && err != goredis.Nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
		}
		span.End()
	}()
	err = fn(ctx)
}

func (c *otelRedisConn) Ping(ctx context.Context) *goredis.StatusCmd {
	var cmd *goredis.StatusCmd
	c.withSpan(ctx, "PING", func(ctx context.Context) error {
		cmd = c.conn.Ping(ctx)
		return cmd.Err()
	})
	return cmd
}

func (c *otelRedisConn) Del(ctx context.Context, keys ...string) *goredis.IntCmd {
	var cmd *goredis.IntCmd
	c.withSpan(ctx, "DEL", func(ctx context.Context) error {
		cmd = c.conn.Del(ctx, keys...)
		return cmd.Err()
	})
	return cmd
}

func (c *otelRedisConn) Get(ctx context.Context, key string) *goredis.StringCmd {
	var cmd *goredis.StringCmd
	c.withSpan(ctx, "GET", func(ctx context.Context) error {
		cmd = c.conn.Get(ctx, key)
		return cmd.Err()
	})
	return cmd
}

func (c *otelRedisConn) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *goredis.StatusCmd {
	var cmd *goredis.StatusCmd
	c.withSpan(ctx, "SET", func(ctx context.Context) error {
		cmd = c.conn.Set(ctx, key, value, expiration)
		return cmd.Err()
	})
	return cmd
}

func (c *otelRedisConn) SetEx(ctx context.Context, key string, value interface{}, expiration time.Duration) *goredis.StatusCmd {
	var cmd *goredis.StatusCmd
	c.withSpan(ctx, "SETEX", func(ctx context.Context) error {
		cmd = c.conn.SetEx(ctx, key, value, expiration)
		return cmd.Err()
	})
	return cmd
}

func (c *otelRedisConn) SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) *goredis.BoolCmd {
	var cmd *goredis.BoolCmd
	c.withSpan(ctx, "SETNX", func(ctx context.Context) error {
		cmd = c.conn.SetNX(ctx, key, value, expiration)
		return cmd.Err()
	})
	return cmd
}

func (c *otelRedisConn) SetXX(ctx context.Context, key string, value interface{}, expiration time.Duration) *goredis.BoolCmd {
	var cmd *goredis.BoolCmd
	c.withSpan(ctx, "SETXX", func(ctx context.Context) error {
		cmd = c.conn.SetXX(ctx, key, value, expiration)
		return cmd.Err()
	})
	return cmd
}

func (c *otelRedisConn) Expire(ctx context.Context, key string, expiration time.Duration) *goredis.BoolCmd {
	var cmd *goredis.BoolCmd
	c.withSpan(ctx, "EXPIRE", func(ctx context.Context) error {
		cmd = c.conn.Expire(ctx, key, expiration)
		return cmd.Err()
	})
	return cmd
}

func (c *otelRedisConn) ExpireAt(ctx context.Context, key string, tm time.Time) *goredis.BoolCmd {
	var cmd *goredis.BoolCmd
	c.withSpan(ctx, "EXPIREAT", func(ctx context.Context) error {
		cmd = c.conn.ExpireAt(ctx, key, tm)
		return cmd.Err()
	})
	return cmd
}

func (c *otelRedisConn) Incr(ctx context.Context, key string) *goredis.IntCmd {
	var cmd *goredis.IntCmd
	c.withSpan(ctx, "INCR", func(ctx context.Context) error {
		cmd = c.conn.Incr(ctx, key)
		return cmd.Err()
	})
	return cmd
}

func (c *otelRedisConn) IncrBy(ctx context.Context, key string, value int64) *goredis.IntCmd {
	var cmd *goredis.IntCmd
	c.withSpan(ctx, "INCRBY", func(ctx context.Context) error {
		cmd = c.conn.IncrBy(ctx, key, value)
		return cmd.Err()
	})
	return cmd
}

func (c *otelRedisConn) PExpireAt(ctx context.Context, key string, tm time.Time) *goredis.BoolCmd {
	var cmd *goredis.BoolCmd
	c.withSpan(ctx, "PEXPIREAT", func(ctx context.Context) error {
		cmd = c.conn.PExpireAt(ctx, key, tm)
		return cmd.Err()
	})
	return cmd
}

func (c *otelRedisConn) XAdd(ctx context.Context, a *goredis.XAddArgs) *goredis.StringCmd {
	var cmd *goredis.StringCmd
	c.withSpan(ctx, "XADD", func(ctx context.Context) error {
		cmd = c.conn.XAdd(ctx, a)
		return cmd.Err()
	})
	return cmd
}

func (c *otelRedisConn) HDel(ctx context.Context, key string, fields ...string) *goredis.IntCmd {
	var cmd *goredis.IntCmd
	c.withSpan(ctx, "HDEL", func(ctx context.Context) error {
		cmd = c.conn.HDel(ctx, key, fields...)
		return cmd.Err()
	})
	return cmd
}

func (c *otelRedisConn) HSet(ctx context.Context, key string, values ...interface{}) *goredis.IntCmd {
	var cmd *goredis.IntCmd
	c.withSpan(ctx, "HSET", func(ctx context.Context) error {
		cmd = c.conn.HSet(ctx, key, values...)
		return cmd.Err()
	})
	return cmd
}

func (c *otelRedisConn) HGetAll(ctx context.Context, key string) *goredis.MapStringStringCmd {
	var cmd *goredis.MapStringStringCmd
	c.withSpan(ctx, "HGETALL", func(ctx context.Context) error {
		cmd = c.conn.HGetAll(ctx, key)
		return cmd.Err()
	})
	return cmd
}

func (c *otelRedisConn) LPush(ctx context.Context, key string, values ...interface{}) *goredis.IntCmd {
	var cmd *goredis.IntCmd
	c.withSpan(ctx, "LPUSH", func(ctx context.Context) error {
		cmd = c.conn.LPush(ctx, key, values...)
		return cmd.Err()
	})
	return cmd
}

func (c *otelRedisConn) BRPop(ctx context.Context, timeout time.Duration, keys ...string) *goredis.StringSliceCmd {
	var cmd *goredis.StringSliceCmd
	c.withSpan(ctx, "BRPOP", func(ctx context.Context) error {
		cmd = c.conn.BRPop(ctx, timeout, keys...)
		return cmd.Err()
	})
	return cmd
}

func (c *otelRedisConn) PFCount(ctx context.Context, keys ...string) *goredis.IntCmd {
	var cmd *goredis.IntCmd
	c.withSpan(ctx, "PFCOUNT", func(ctx context.Context) error {
		cmd = c.conn.PFCount(ctx, keys...)
		return cmd.Err()
	})
	return cmd
}

func (c *otelRedisConn) PFAdd(ctx context.Context, key string, els ...interface{}) *goredis.IntCmd {
	var cmd *goredis.IntCmd
	c.withSpan(ctx, "PFADD", func(ctx context.Context) error {
		cmd = c.conn.PFAdd(ctx, key, els...)
		return cmd.Err()
	})
	return cmd
}

func (c *otelRedisConn) Eval(ctx context.Context, script string, keys []string, args ...interface{}) *goredis.Cmd {
	var cmd *goredis.Cmd
	c.withSpan(ctx, "EVAL", func(ctx context.Context) error {
		cmd = c.conn.Eval(ctx, script, keys, args...)
		return cmd.Err()
	})
	return cmd
}

func (c *otelRedisConn) EvalSha(ctx context.Context, sha1 string, keys []string, args ...interface{}) *goredis.Cmd {
	var cmd *goredis.Cmd
	c.withSpan(ctx, "EVALSHA", func(ctx context.Context) error {
		cmd = c.conn.EvalSha(ctx, sha1, keys, args...)
		return cmd.Err()
	})
	return cmd
}

func (c *otelRedisConn) EvalRO(ctx context.Context, script string, keys []string, args ...interface{}) *goredis.Cmd {
	var cmd *goredis.Cmd
	c.withSpan(ctx, "EVAL_RO", func(ctx context.Context) error {
		cmd = c.conn.EvalRO(ctx, script, keys, args...)
		return cmd.Err()
	})
	return cmd
}

func (c *otelRedisConn) EvalShaRO(ctx context.Context, sha1 string, keys []string, args ...interface{}) *goredis.Cmd {
	var cmd *goredis.Cmd
	c.withSpan(ctx, "EVALSHA_RO", func(ctx context.Context) error {
		cmd = c.conn.EvalShaRO(ctx, sha1, keys, args...)
		return cmd.Err()
	})
	return cmd
}

func (c *otelRedisConn) ScriptExists(ctx context.Context, hashes ...string) *goredis.BoolSliceCmd {
	var cmd *goredis.BoolSliceCmd
	c.withSpan(ctx, "SCRIPT_EXISTS", func(ctx context.Context) error {
		cmd = c.conn.ScriptExists(ctx, hashes...)
		return cmd.Err()
	})
	return cmd
}

func (c *otelRedisConn) ScriptLoad(ctx context.Context, script string) *goredis.StringCmd {
	var cmd *goredis.StringCmd
	c.withSpan(ctx, "SCRIPT_LOAD", func(ctx context.Context) error {
		cmd = c.conn.ScriptLoad(ctx, script)
		return cmd.Err()
	})
	return cmd
}
