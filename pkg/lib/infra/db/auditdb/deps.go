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
)

type SQLBuilder struct {
	db.SQLBuilder
}

func NewSQLBuilder(c *config.AuditDatabaseCredentials, id config.AppID) *SQLBuilder {
	if c == nil {
		return nil
	}

	return &SQLBuilder{
		db.NewSQLBuilder(c.DatabaseSchema, string(id)),
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
	cfg *config.DatabaseConfig,
	credentials *config.AuditDatabaseCredentials,
	lf *log.Factory,
) *ReadHandle {
	if credentials == nil {
		return nil
	}

	opts := db.ConnectionOptions{
		DatabaseURL:           credentials.DatabaseURL,
		MaxOpenConnection:     *cfg.MaxOpenConnection,
		MaxIdleConnection:     *cfg.MaxIdleConnection,
		MaxConnectionLifetime: cfg.MaxConnectionLifetime.Duration(),
		IdleConnectionTimeout: cfg.IdleConnectionTimeout.Duration(),
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
	cfg *config.DatabaseConfig,
	credentials *config.AuditDatabaseCredentials,
	lf *log.Factory,
) *WriteHandle {
	if credentials == nil {
		return nil
	}

	opts := db.ConnectionOptions{
		DatabaseURL:           credentials.DatabaseURL,
		MaxOpenConnection:     *cfg.MaxOpenConnection,
		MaxIdleConnection:     *cfg.MaxIdleConnection,
		MaxConnectionLifetime: cfg.MaxConnectionLifetime.Duration(),
		IdleConnectionTimeout: cfg.IdleConnectionTimeout.Duration(),
	}
	return &WriteHandle{
		db.NewHookHandle(ctx, pool, opts, lf),
	}
}

func (h *WriteHandle) WithTx(do func() error) (err error) {
	return h.HookHandle.WithTx(do)
}
