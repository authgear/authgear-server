package redis

import (
	redigo "github.com/gomodule/redigo/redis"

	"github.com/skygeario/skygear-server/pkg/auth/config"
	"github.com/skygeario/skygear-server/pkg/log"
)

type Conn = redigo.Conn

type Context struct {
	pool        *Pool
	cfg         *config.RedisConfig
	credentials *config.RedisCredentials
	logger      *log.Logger
}

func NewContext(pool *Pool, cfg *config.RedisConfig, credentials *config.RedisCredentials, lf *log.Factory) *Context {
	return &Context{
		pool:        pool,
		cfg:         cfg,
		logger:      lf.New("rediscontext"),
		credentials: credentials,
	}
}

func (ctx *Context) WithConn(f func(conn Conn) error) error {
	ctx.logger.WithFields(map[string]interface{}{
		"max_open_connection":             *ctx.cfg.MaxOpenConnection,
		"max_idle_connection":             *ctx.cfg.MaxIdleConnection,
		"idle_connection_timeout_seconds": *ctx.cfg.IdleConnectionTimeout,
		"max_connection_lifetime_seconds": *ctx.cfg.MaxConnectionLifetime,
	}).Debug("open redis connection")

	conn := ctx.pool.Open(ctx.cfg, ctx.credentials).Get()
	defer func() {
		err := conn.Close()
		if err != nil {
			ctx.logger.WithError(err).Error("failed to close connection")
		}
	}()

	return f(conn)
}
