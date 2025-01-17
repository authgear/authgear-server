package globaldb

import (
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

func NewSQLExecutor(handle *Handle) *SQLExecutor {
	return &SQLExecutor{
		db.SQLExecutor{},
	}
}

type Handle struct {
	*db.HookHandle
}

func NewHandle(
	pool *db.Pool,
	credentials *config.GlobalDatabaseCredentialsEnvironmentConfig,
	cfg *config.DatabaseEnvironmentConfig,
	lf *log.Factory,
) *Handle {
	info := db.ConnectionInfo{
		Purpose:     db.ConnectionPurposeGlobal,
		DatabaseURL: credentials.DatabaseURL,
	}

	opts := db.ConnectionOptions{
		MaxOpenConnection:     cfg.MaxOpenConn,
		MaxIdleConnection:     cfg.MaxIdleConn,
		MaxConnectionLifetime: time.Second * time.Duration(cfg.ConnMaxLifetimeSeconds),
		IdleConnectionTimeout: time.Second * time.Duration(cfg.ConnMaxIdleTimeSeconds),
	}
	return &Handle{
		db.NewHookHandle(pool, info, opts, lf),
	}
}
