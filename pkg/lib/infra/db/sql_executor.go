package db

import (
	"context"
	"database/sql"
	"errors"

	sq "github.com/Masterminds/squirrel"
	"github.com/lib/pq"

	"github.com/authgear/authgear-server/pkg/util/errorutil"
)

type SQLExecutor struct {
	Database Handle
}

func (e *SQLExecutor) ExecWith(ctx context.Context, sqlizeri sq.Sqlizer) (sql.Result, error) {
	db, err := e.Database.conn()
	if err != nil {
		return nil, err
	}
	sql, args, err := sqlizeri.ToSql()
	if err != nil {
		return nil, err
	}
	result, err := db.ExecContext(ctx, sql, args...)
	if err != nil {
		if isWriteConflict(err) {
			panic(ErrWriteConflict)
		}
		return nil, errorutil.WithDetails(err, errorutil.Details{"sql": errorutil.SafeDetail.Value(sql)})
	}
	return result, nil
}

func (e *SQLExecutor) QueryWith(ctx context.Context, sqlizeri sq.Sqlizer) (*sql.Rows, error) {
	db, err := e.Database.conn()
	if err != nil {
		return nil, err
	}
	sql, args, err := sqlizeri.ToSql()
	if err != nil {
		return nil, err
	}
	result, err := db.QueryxContext(ctx, sql, args...)
	if err != nil {
		if isWriteConflict(err) {
			panic(ErrWriteConflict)
		}
		return nil, errorutil.WithDetails(err, errorutil.Details{"sql": errorutil.SafeDetail.Value(sql)})
	}
	return result, nil
}

func (e *SQLExecutor) QueryRowWith(ctx context.Context, sqlizeri sq.Sqlizer) (*sql.Row, error) {
	db, err := e.Database.conn()
	if err != nil {
		return nil, err
	}
	sql, args, err := sqlizeri.ToSql()
	if err != nil {
		if isWriteConflict(err) {
			panic(ErrWriteConflict)
		}
		return nil, errorutil.WithDetails(err, errorutil.Details{"sql": errorutil.SafeDetail.Value(sql)})
	}
	return db.QueryRowxContext(ctx, sql, args...), nil
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
