package util

import (
	"context"
	"log/slog"
	"time"

	"github.com/lib/pq"

	"github.com/authgear/authgear-server/pkg/lib/infra/db"
	"github.com/authgear/authgear-server/pkg/util/slogutil"
)

var PrunerLogger = slogutil.NewLogger("pruner")

type Pruner struct {
	ConnectionInfo db.ConnectionInfo
	DatabaseSchema string
	AppIDs         []string
	TableNames     []string
	DryRun         bool

	dbHandle    *db.HookHandle
	sqlExecutor *db.SQLExecutor
	sqlBuilder  *db.SQLBuilder
}

func NewPruner(
	connectionInfo db.ConnectionInfo,
	databaseSchema string,
	appIDs []string,
	tableNames []string,
	dryRun bool,
) *Pruner {
	pool := db.NewPool()
	handle := db.NewHookHandle(
		pool,
		connectionInfo,
		db.ConnectionOptions{
			MaxOpenConnection:     1,
			MaxIdleConnection:     1,
			MaxConnectionLifetime: 1800 * time.Second,
			IdleConnectionTimeout: 300 * time.Second,
		},
	)
	sqlExecutor := &db.SQLExecutor{}
	sqlBuilder := db.NewSQLBuilder(databaseSchema)
	return &Pruner{
		ConnectionInfo: connectionInfo,
		DatabaseSchema: databaseSchema,
		AppIDs:         appIDs,
		TableNames:     tableNames,
		DryRun:         dryRun,

		dbHandle:    handle,
		sqlExecutor: sqlExecutor,
		sqlBuilder:  &sqlBuilder,
	}
}

// Prune deletes every row keyed by p.AppIDs from every table in p.TableNames.
// TableNames must be ordered so that a referencing table precedes any table
// it references (i.e. children before parents), otherwise foreign key
// constraints will reject the delete.
func (p *Pruner) Prune(ctx context.Context) error {
	if p.DryRun {
		return p.dbHandle.ReadOnly(ctx, func(ctx context.Context) error {
			logger := PrunerLogger.GetLogger(ctx)
			for _, tableName := range p.TableNames {
				count, err := p.countTable(ctx, tableName)
				if err != nil {
					return err
				}
				logger.Info(ctx, "would delete rows", slog.String("table", tableName), slog.Int64("count", count))
			}
			return nil
		})
	}

	return p.dbHandle.WithTx(ctx, func(ctx context.Context) error {
		logger := PrunerLogger.GetLogger(ctx)
		for _, tableName := range p.TableNames {
			affected, err := p.deleteTable(ctx, tableName)
			if err != nil {
				return err
			}
			logger.Info(ctx, "deleted rows", slog.String("table", tableName), slog.Int64("count", affected))
		}
		return nil
	})
}

func (p *Pruner) countTable(ctx context.Context, tableName string) (int64, error) {
	q := p.sqlBuilder.Select("COUNT(*)").
		From(p.sqlBuilder.TableName(tableName)).
		Where("app_id = ANY (?)", pq.Array(p.AppIDs))

	row, err := p.sqlExecutor.QueryRowWith(ctx, q)
	if err != nil {
		return 0, err
	}
	var count int64
	if err := row.Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

func (p *Pruner) deleteTable(ctx context.Context, tableName string) (int64, error) {
	q := p.sqlBuilder.Delete(p.sqlBuilder.TableName(tableName)).
		Where("app_id = ANY (?)", pq.Array(p.AppIDs))

	result, err := p.sqlExecutor.ExecWith(ctx, q)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}
