package redis

import (
	"context"

	"github.com/gomodule/redigo/redis"
)

type contextKeyType struct{}

var contextKey = contextKeyType{}

type redisContext struct {
	pool *redis.Pool
	conn redis.Conn
}

func WithRedis(ctx context.Context, pool *redis.Pool) context.Context {
	redisCtx := &redisContext{pool: pool, conn: nil}
	return context.WithValue(ctx, contextKey, redisCtx)
}

func GetConn(ctx context.Context) redis.Conn {
	redisCtx := ctx.Value(contextKey).(*redisContext)
	if redisCtx.conn == nil {
		redisCtx.conn = redisCtx.pool.Get()
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
