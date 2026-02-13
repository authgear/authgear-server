package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"strings"

	sq "github.com/Masterminds/squirrel"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	semconv "go.opentelemetry.io/otel/semconv/v1.34.0"
	"go.opentelemetry.io/otel/trace"

	"github.com/lib/pq"

	"github.com/authgear/authgear-server/pkg/util/errorutil"
)

var tracer = otel.Tracer("github.com/authgear/authgear-server/pkg/lib/infra/db")

func getOperationName(sql string) string {
	parts := strings.Fields(sql)
	if len(parts) > 0 {
		return strings.ToUpper(parts[0])
	}
	return "UNKNOWN"
}

type SQLExecutor struct{}

func (e *SQLExecutor) withSpan(ctx context.Context, sql string, fn func(ctx context.Context) error) error {
	operation := getOperationName(sql)
	ctx, span := tracer.Start(ctx, "DB "+operation, trace.WithAttributes(
		attribute.String("db.system", "postgresql"),
		semconv.DBOperationName(operation),
		semconv.DBQueryText(sql),
	))
	var err error
	defer func() {
		if r := recover(); r != nil {
			err := fmt.Errorf("panic: %v", r)
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			span.End()
			panic(r)
		}
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
		}
		span.End()
	}()
	err = fn(ctx)
	return err
}

func (e *SQLExecutor) ExecWith(ctx context.Context, sqlizeri sq.Sqlizer) (sql.Result, error) {
	var result sql.Result
	sql, args, err := sqlizeri.ToSql()
	if err != nil {
		return nil, err
	}

	err = e.withSpan(ctx, sql, func(ctx context.Context) error {
		stmtPreparer, ok := getStmtPreparer(ctx)
		if !ok {
			tx := mustGetTxLike(ctx)
			res, err := tx.ExecContext(ctx, sql, args...)
			if err != nil {
				if isWriteConflict(err) {
					panic(ErrWriteConflict)
				}
				return errorutil.WithDetails(err, errorutil.Details{"sql": errorutil.SafeDetail.Value(sql)})
			}
			result = res
			return nil
		}

		stmt, err := stmtPreparer.PrepareContext(ctx, sql)
		if err != nil {
			return err
		}

		res, err := stmt.ExecContext(ctx, args...)
		if err != nil {
			if isWriteConflict(err) {
				panic(ErrWriteConflict)
			}
			return errorutil.WithDetails(err, errorutil.Details{"sql": errorutil.SafeDetail.Value(sql)})
		}
		result = res
		return nil
	})
	return result, err
}

func (e *SQLExecutor) QueryWith(ctx context.Context, sqlizeri sq.Sqlizer) (*sql.Rows, error) {
	var result *sql.Rows
	sql, args, err := sqlizeri.ToSql()
	if err != nil {
		return nil, err
	}

	err = e.withSpan(ctx, sql, func(ctx context.Context) error {
		stmtPreparer, ok := getStmtPreparer(ctx)
		if !ok {
			tx := mustGetTxLike(ctx)
			res, err := tx.QueryContext(ctx, sql, args...)
			if err != nil {
				if isWriteConflict(err) {
					panic(ErrWriteConflict)
				}
				return errorutil.WithDetails(err, errorutil.Details{"sql": errorutil.SafeDetail.Value(sql)})
			}
			result = res
			return nil
		}

		stmt, err := stmtPreparer.PrepareContext(ctx, sql)
		if err != nil {
			return err
		}

		res, err := stmt.QueryContext(ctx, args...)
		if err != nil {
			if isWriteConflict(err) {
				panic(ErrWriteConflict)
			}
			return errorutil.WithDetails(err, errorutil.Details{"sql": errorutil.SafeDetail.Value(sql)})
		}
		result = res
		return nil
	})
	return result, err
}

func (e *SQLExecutor) QueryRowWith(ctx context.Context, sqlizeri sq.Sqlizer) (*sql.Row, error) {
	var result *sql.Row
	sql, args, err := sqlizeri.ToSql()
	if err != nil {
		return nil, err
	}

	e.withSpan(ctx, sql, func(ctx context.Context) error {
		stmtPreparer, ok := getStmtPreparer(ctx)
		if !ok {
			tx := mustGetTxLike(ctx)
			result = tx.QueryRowContext(ctx, sql, args...)
			return nil
		}

		stmt, err := stmtPreparer.PrepareContext(ctx, sql)
		if err != nil {
			return err
		}

		result = stmt.QueryRowContext(ctx, args...)
		return nil
	})
	return result, nil
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
