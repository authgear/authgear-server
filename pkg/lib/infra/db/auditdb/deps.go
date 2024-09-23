package auditdb

import (
	"context"

	"github.com/google/wire"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/db"
	"github.com/authgear/authgear-server/pkg/util/log"
)

var DependencySet = wire.NewSet(
	NewReadSQLExecutor,
	NewWriteSQLExecutor,
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

func NewSQLBuilder(c *config.AuditDatabaseCredentials) *SQLBuilder {
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

func NewSQLBuilderApp(c *config.AuditDatabaseCredentials, id config.AppID) *SQLBuilderApp {
	if c == nil {
		return nil
	}

	return &SQLBuilderApp{
		db.NewSQLBuilderApp(c.DatabaseSchema, string(id)),
	}
}

type ReadSQLExecutor struct {
	db.SQLExecutor
}

type WriteSQLExecutor struct {
	db.SQLExecutor
}

func NewReadSQLExecutor(c context.Context, handle *ReadHandle) *ReadSQLExecutor {
	if handle == nil {
		return nil
	}

	return &ReadSQLExecutor{
		db.SQLExecutor{
			Context:  c,
			Database: handle,
		},
	}
}

func NewWriteSQLExecutor(c context.Context, handle *WriteHandle) *WriteSQLExecutor {
	if handle == nil {
		return nil
	}

	return &WriteSQLExecutor{
		db.SQLExecutor{
			Context:  c,
			Database: handle,
		},
	}
}

type ReadHandle struct {
	*db.HookHandle
}

func NewReadHandle(
	ctx context.Context,
	pool *db.Pool,
	cfg *config.DatabaseEnvironmentConfig,
	credentials *config.AuditDatabaseCredentials,
	lf *log.Factory,
) *ReadHandle {
	if credentials == nil {
		return nil
	}

	opts := db.ConnectionOptions{
		DatabaseURL:           credentials.DatabaseURL,
		MaxOpenConnection:     cfg.MaxOpenConn,
		MaxIdleConnection:     cfg.MaxIdleConn,
		MaxConnectionLifetime: cfg.ConnMaxLifetimeSeconds.Duration(),
		IdleConnectionTimeout: cfg.ConnMaxIdleTimeSeconds.Duration(),
		UsePreparedStatements: cfg.UsePreparedStatements,
	}
	return &ReadHandle{
		db.NewHookHandle(ctx, pool, opts, lf),
	}
}

func (h *ReadHandle) ReadOnly(do func() error) (err error) {
	return h.HookHandle.ReadOnly(do)
}

type WriteHandle struct {
	*db.HookHandle
}

func NewWriteHandle(
	ctx context.Context,
	pool *db.Pool,
	cfg *config.DatabaseEnvironmentConfig,
	credentials *config.AuditDatabaseCredentials,
	lf *log.Factory,
) *WriteHandle {
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
	return &WriteHandle{
		db.NewHookHandle(ctx, pool, opts, lf),
	}
}

func (h *WriteHandle) WithTx(do func() error) (err error) {
	return h.HookHandle.WithTx(do)
}
