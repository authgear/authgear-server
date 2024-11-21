package db

import (
	"context"
	"database/sql"
	"errors"

	sq "github.com/Masterminds/squirrel"
	"github.com/lib/pq"

	"github.com/authgear/authgear-server/pkg/util/errorutil"
)

type SQLExecutor struct{}

func (e *SQLExecutor) ExecWith(ctx context.Context, sqlizeri sq.Sqlizer) (sql.Result, error) {
	sql, args, err := sqlizeri.ToSql()
	if err != nil {
		return nil, err
	}

	stmtPreparer, ok := getStmtPreparer(ctx)
	if !ok {
		tx := mustGetTxLike(ctx)
		result, err := tx.ExecContext(ctx, sql, args...)
		if err != nil {
			if isWriteConflict(err) {
				panic(ErrWriteConflict)
			}
			return nil, errorutil.WithDetails(err, errorutil.Details{"sql": errorutil.SafeDetail.Value(sql)})
		}
		return result, nil
	}

	stmt, err := stmtPreparer.PrepareContext(ctx, sql)
	if err != nil {
		return nil, err
	}

	result, err := stmt.ExecContext(ctx, args...)
	if err != nil {
		if isWriteConflict(err) {
			panic(ErrWriteConflict)
		}
		return nil, errorutil.WithDetails(err, errorutil.Details{"sql": errorutil.SafeDetail.Value(sql)})
	}
	return result, nil
}

func (e *SQLExecutor) QueryWith(ctx context.Context, sqlizeri sq.Sqlizer) (*sql.Rows, error) {
	sql, args, err := sqlizeri.ToSql()
	if err != nil {
		return nil, err
	}

	stmtPreparer, ok := getStmtPreparer(ctx)
	if !ok {
		tx := mustGetTxLike(ctx)
		result, err := tx.QueryContext(ctx, sql, args...)
		if err != nil {
			if isWriteConflict(err) {
				panic(ErrWriteConflict)
			}
			return nil, errorutil.WithDetails(err, errorutil.Details{"sql": errorutil.SafeDetail.Value(sql)})
		}
		return result, nil
	}

	stmt, err := stmtPreparer.PrepareContext(ctx, sql)
	if err != nil {
		return nil, err
	}

	result, err := stmt.QueryContext(ctx, args...)
	if err != nil {
		if isWriteConflict(err) {
			panic(ErrWriteConflict)
		}
		return nil, errorutil.WithDetails(err, errorutil.Details{"sql": errorutil.SafeDetail.Value(sql)})
	}
	return result, nil
}

func (e *SQLExecutor) QueryRowWith(ctx context.Context, sqlizeri sq.Sqlizer) (*sql.Row, error) {
	sql, args, err := sqlizeri.ToSql()
	if err != nil {
		return nil, err
	}

	stmtPreparer, ok := getStmtPreparer(ctx)
	if !ok {
		tx := mustGetTxLike(ctx)
		return tx.QueryRowContext(ctx, sql, args...), nil
	}

	stmt, err := stmtPreparer.PrepareContext(ctx, sql)
	if err != nil {
		return nil, err
	}

	return stmt.QueryRowContext(ctx, args...), nil
}

func isWriteConflict(err error) bool {
	var pqErr *pq.Error
	if errors.As(err, &pqErr) {
		// 40001: serialization_failure
		// 40P01: deadlock_detected
		return pqErr.Code == "40001" || pqErr.Code == "40P01"
	}
	return false
}
