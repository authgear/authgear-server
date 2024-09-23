package db

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"

	"github.com/authgear/authgear-server/pkg/util/log"
)

type txConn struct {
	doPrepare bool
	logger    *log.Logger
	tx        *sqlx.Tx
	db        *PoolDB
}

func (t *txConn) prepare(ctx context.Context, query string) (*sqlx.Stmt, error) {
	stmt, err := t.db.Prepare(ctx, query)
	if err != nil {
		return nil, err
	}

	return t.tx.Stmtx(stmt), nil
}

func (t *txConn) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	if !t.doPrepare {
		rows, err := t.tx.QueryContext(ctx, query, args...)
		if err != nil {
			t.logger.WithError(err).Debug("failed to execute query")
		}
		return rows, err
	}

	stmt, err := t.prepare(ctx, query)
	if err != nil {
		t.logger.WithError(err).WithField("query", query).Error("failed to prepare statement")
		return t.tx.QueryContext(ctx, query, args...)
	}

	rows, err := stmt.QueryContext(ctx, args...)
	if err != nil {
		t.logger.WithError(err).Debug("failed to execute prepared statement")
	}
	return rows, err
}

func (t *txConn) QueryRowxContext(ctx context.Context, query string, args ...interface{}) *sqlx.Row {
	if !t.doPrepare {
		row := t.tx.QueryRowxContext(ctx, query, args...)
		if err := row.Err(); err != nil {
			t.logger.WithError(err).Debug("failed to execute query")
		}
		return row
	}

	stmt, err := t.prepare(ctx, query)
	if err != nil {
		t.logger.WithError(err).WithField("query", query).Error("failed to prepare statement")
		return t.tx.QueryRowxContext(ctx, query, args...)
	}

	row := stmt.QueryRowxContext(ctx, args...)
	if err := row.Err(); err != nil {
		t.logger.WithError(err).Debug("failed to execute prepared statement")
	}
	return row
}

func (t *txConn) QueryxContext(ctx context.Context, query string, args ...interface{}) (*sqlx.Rows, error) {
	if !t.doPrepare {
		rows, err := t.tx.QueryxContext(ctx, query, args...)
		if err != nil {
			t.logger.WithError(err).Debug("failed to execute query")
		}
		return rows, err
	}

	stmt, err := t.prepare(ctx, query)
	if err != nil {
		t.logger.WithError(err).WithField("query", query).Error("failed to prepare statement")
		return t.tx.QueryxContext(ctx, query, args...)
	}

	rows, err := stmt.QueryxContext(ctx, args...)
	if err != nil {
		t.logger.WithError(err).Debug("failed to execute prepared statement")
	}
	return rows, err
}

func (t *txConn) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	if !t.doPrepare {
		result, err := t.tx.ExecContext(ctx, query, args...)
		if err != nil {
			t.logger.WithError(err).Debug("failed to execute query")
		}
		return result, err
	}

	stmt, err := t.prepare(ctx, query)
	if err != nil {
		t.logger.WithError(err).WithField("query", query).Error("failed to prepare statement")
		return t.tx.ExecContext(ctx, query, args...)
	}

	result, err := stmt.ExecContext(ctx, args...)
	if err != nil {
		t.logger.WithError(err).Debug("failed to execute prepared statement")
	}
	return result, err
}
