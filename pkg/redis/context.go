package redis

import (
	redigo "github.com/gomodule/redigo/redis"

	"github.com/skygeario/skygear-server/pkg/auth/config"
	"github.com/skygeario/skygear-server/pkg/log"
)

type Context struct {
	pool        *Pool
	cfg         *config.RedisConfig
	credentials *config.RedisCredentials
	logger      *log.Logger

	conn redigo.Conn
}

func NewContext(pool *Pool, cfg *config.RedisConfig, credentials *config.RedisCredentials, lf *log.Factory) *Context {
	return &Context{
		pool:        pool,
		cfg:         cfg,
		logger:      lf.New("rediscontext"),
		credentials: credentials,
		conn:        nil,
	}
}

func (ctx *Context) Conn() redigo.Conn {
	if ctx.conn == nil {
		ctx.logger.WithFields(map[string]interface{}{
			"max_open_connection":             *ctx.cfg.MaxOpenConnection,
			"max_idle_connection":             *ctx.cfg.MaxIdleConnection,
			"idle_connection_timeout_seconds": *ctx.cfg.IdleConnectionTimeout,
			"max_connection_lifetime_seconds": *ctx.cfg.MaxConnectionLifetime,
		}).Debug("open redis connection")

		ctx.conn = ctx.pool.Open(ctx.cfg, ctx.credentials).Get()
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
