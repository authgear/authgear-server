package redis

import (
	redigo "github.com/gomodule/redigo/redis"
	"github.com/skygeario/skygear-server/pkg/auth/config"
)

type Context struct {
	pool        *Pool
	credentials *config.RedisCredentials
	conn        redigo.Conn
}

func NewContext(pool *Pool, c *config.RedisCredentials) *Context {
	return &Context{pool: pool, credentials: c, conn: nil}
}

func (ctx *Context) Conn() redigo.Conn {
	if ctx.conn == nil {
		ctx.conn = ctx.pool.Open(ctx.credentials).Get()
	}
	return ctx.conn
}

func (ctx *Context) Close() error {
	if ctx.conn == nil {
		return nil
	}
	conn := ctx.conn
	ctx.conn = nil
	return conn.Close()
}
