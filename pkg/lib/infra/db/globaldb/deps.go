package globaldb

import (
	"context"
	"time"

	"github.com/google/wire"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/db"
	"github.com/authgear/authgear-server/pkg/util/log"
)

var DependencySet = wire.NewSet(
	NewHandle,
	NewSQLExecutor,
	NewSQLBuilder,
)

type SQLBuilder struct {
	db.SQLBuilder
}

func NewSQLBuilder(
	credentials *config.GlobalDatabaseCredentialsEnvironmentConfig,
) *SQLBuilder {
	return &SQLBuilder{
		db.NewSQLBuilder(credentials.DatabaseSchema),
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
	credentials *config.GlobalDatabaseCredentialsEnvironmentConfig,
	cfg *config.DatabaseEnvironmentConfig,
	lf *log.Factory,
) *Handle {
	opts := db.ConnectionOptions{
		DatabaseURL:           credentials.DatabaseURL,
		MaxOpenConnection:     cfg.MaxOpenConn,
		MaxIdleConnection:     cfg.MaxIdleConn,
		MaxConnectionLifetime: time.Second * time.Duration(cfg.ConnMaxLifetimeSeconds),
		IdleConnectionTimeout: time.Second * time.Duration(cfg.ConnMaxIdleTimeSeconds),
		UsePreparedStatements: cfg.UsePreparedStatements,
	}
	return &Handle{
		db.NewHookHandle(ctx, pool, opts, lf),
	}
}
