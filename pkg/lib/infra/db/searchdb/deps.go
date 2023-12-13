package searchdb

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
	NewSQLBuilderApp,
)

type SQLBuilder struct {
	builder *db.SQLBuilder
}

func (b SQLBuilder) WithoutAppID() *db.SQLBuilder {
	return b.builder
}

func (b SQLBuilder) WithAppID(appID string) *SQLBuilderApp {
	return &SQLBuilderApp{
		SQLBuilderApp: db.NewSQLBuilderApp(b.builder.Schema, appID),
	}
}

func (b SQLBuilder) TableName(table string) string {
	return b.builder.TableName(table)
}

func NewSQLBuilder(c *config.SearchDatabaseCredentials) *SQLBuilder {
	if c == nil {
		return nil
	}

	builder := db.NewSQLBuilder(c.DatabaseSchema)
	return &SQLBuilder{
		builder: &builder,
	}
}

type SQLBuilderApp struct {
	db.SQLBuilderApp
}

func NewSQLBuilderApp(c *config.SearchDatabaseCredentials, id config.AppID) *SQLBuilderApp {
	if c == nil {
		return nil
	}

	return &SQLBuilderApp{
		db.NewSQLBuilderApp(c.DatabaseSchema, string(id)),
	}
}

type SQLExecutor struct {
	db.SQLExecutor
}

func NewSQLExecutor(c context.Context, handle *Handle) *SQLExecutor {
	if handle == nil {
		return nil
	}

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
	cfg *config.DatabaseEnvironmentConfig,
	credentials *config.SearchDatabaseCredentials,
	lf *log.Factory,
) *Handle {
	if credentials == nil {
		return nil
	}
	opts := db.ConnectionOptions{
		DatabaseURL:           credentials.DatabaseURL,
		MaxOpenConnection:     cfg.MaxOpenConn,
		MaxIdleConnection:     cfg.MaxIdleConn,
		MaxConnectionLifetime: cfg.ConnMaxLifetimeSeconds.Duration(),
		IdleConnectionTimeout: cfg.ConnMaxIdleTimeSeconds.Duration(),
	}
	return &Handle{
		db.NewHookHandle(ctx, pool, opts, lf),
	}
}
