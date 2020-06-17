package redis

import (
	"context"
	"github.com/skygeario/skygear-server/pkg/auth/config"

	"github.com/gomodule/redigo/redis"
)

type contextKeyType struct{}

var contextKey = contextKeyType{}

type redisContext struct {
	pool        *Pool
	credentials *config.RedisCredentials
	conn        redis.Conn
}

func WithRedis(ctx context.Context, pool *Pool, c *config.RedisCredentials) context.Context {
	redisCtx := &redisContext{pool: pool, credentials: c, conn: nil}
	return context.WithValue(ctx, contextKey, redisCtx)
}

func GetConn(ctx context.Context) redis.Conn {
	redisCtx := ctx.Value(contextKey).(*redisContext)
	if redisCtx.conn == nil {
		redisCtx.conn = redisCtx.pool.Open(redisCtx.credentials).Get()
	}
	return redisCtx.conn
}

func CloseConn(ctx context.Context) error {
	redisCtx := ctx.Value(contextKey).(*redisContext)
	if redisCtx.conn == nil {
		return nil
	}
	conn := redisCtx.conn
	redisCtx.conn = nil
	return conn.Close()
}
