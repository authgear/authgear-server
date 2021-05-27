package tenant

import (
	"context"

	"github.com/google/wire"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/db"
	"github.com/authgear/authgear-server/pkg/util/log"
)

var DependencySet = wire.NewSet(
	NewSQLExecutor,
	NewSQLBuilder,
)

type SQLBuilder struct {
	db.SQLBuilder
}

func NewSQLBuilder(c *config.DatabaseCredentials, id config.AppID) *SQLBuilder {
	return &SQLBuilder{
		db.NewSQLBuilder(c.DatabaseSchema, string(id)),
	}
}

type SQLExecutor struct {
	db.SQLExecutor
}

func NewSQLExecutor(c context.Context, handle *Handle) *SQLExecutor {
	return &SQLExecutor{
		db.SQLExecutor{
			Context:  c,
			Database: handle,
		},
	}
}

type Handle struct {
	*db.HookHandle
}

func NewHandle(
	ctx context.Context,
	pool *db.Pool,
	cfg *config.DatabaseConfig,
	credentials *config.DatabaseCredentials,
	lf *log.Factory,
) *Handle {
	opts := db.ConnectionOptions{
		DatabaseURL:           credentials.DatabaseURL,
		MaxOpenConnection:     *cfg.MaxOpenConnection,
		MaxIdleConnection:     *cfg.MaxIdleConnection,
		MaxConnectionLifetime: cfg.MaxConnectionLifetime.Duration(),
		IdleConnectionTimeout: cfg.IdleConnectionTimeout.Duration(),
	}
	return &Handle{
		db.NewHookHandle(ctx, pool, opts, lf),
	}
}
