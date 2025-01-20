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

func NewReadSQLExecutor(handle *ReadHandle) *ReadSQLExecutor {
	if handle == nil {
		return nil
	}

	return &ReadSQLExecutor{
		db.SQLExecutor{},
	}
}

func NewWriteSQLExecutor(handle *WriteHandle) *WriteSQLExecutor {
	if handle == nil {
		return nil
	}

	return &WriteSQLExecutor{
		db.SQLExecutor{},
	}
}

type ReadHandle struct {
	*db.HookHandle
}

func NewReadHandle(
	pool *db.Pool,
	cfg *config.DatabaseEnvironmentConfig,
	credentials *config.AuditDatabaseCredentials,
	lf *log.Factory,
) *ReadHandle {
	if credentials == nil {
		return nil
	}

	info := db.ConnectionInfo{
		Purpose:     db.ConnectionPurposeAuditReadOnly,
		DatabaseURL: credentials.DatabaseURL,
	}

	opts := db.ConnectionOptions{
		MaxOpenConnection:     cfg.MaxOpenConn,
		MaxIdleConnection:     cfg.MaxIdleConn,
		MaxConnectionLifetime: cfg.ConnMaxLifetimeSeconds.Duration(),
		IdleConnectionTimeout: cfg.ConnMaxIdleTimeSeconds.Duration(),
	}
	return &ReadHandle{
		db.NewHookHandle(pool, info, opts, lf),
	}
}

func (h *ReadHandle) ReadOnly(ctx context.Context, do func(ctx context.Context) error) (err error) {
	return h.HookHandle.ReadOnly(ctx, do)
}

type WriteHandle struct {
	*db.HookHandle
}

func NewWriteHandle(
	pool *db.Pool,
	cfg *config.DatabaseEnvironmentConfig,
	credentials *config.AuditDatabaseCredentials,
	lf *log.Factory,
) *WriteHandle {
	if credentials == nil {
		return nil
	}

	info := db.ConnectionInfo{
		Purpose:     db.ConnectionPurposeAuditReadWrite,
		DatabaseURL: credentials.DatabaseURL,
	}

	opts := db.ConnectionOptions{
		MaxOpenConnection:     cfg.MaxOpenConn,
		MaxIdleConnection:     cfg.MaxIdleConn,
		MaxConnectionLifetime: cfg.ConnMaxLifetimeSeconds.Duration(),
		IdleConnectionTimeout: cfg.ConnMaxIdleTimeSeconds.Duration(),
	}
	return &WriteHandle{
		db.NewHookHandle(pool, info, opts, lf),
	}
}

func (h *WriteHandle) WithTx(ctx context.Context, do func(ctx context.Context) error) (err error) {
	return h.HookHandle.WithTx(ctx, do)
}
