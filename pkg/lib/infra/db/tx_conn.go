package db

import (
	"context"
	"database/sql"

	"github.com/authgear/authgear-server/pkg/util/log"
)

type txConn struct {
	doPrepare bool
	logger    *log.Logger
	tx        *sql.Tx
	conn      *sql.Conn
}

func (t *txConn) prepare(ctx context.Context, query string) (*sql.Stmt, error) {
	stmt, err := t.conn.PrepareContext(ctx, query)
	if err != nil {
		return nil, err
	}

	return t.tx.Stmt(stmt), nil
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

func (t *txConn) QueryRowxContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	if !t.doPrepare {
		row := t.tx.QueryRowContext(ctx, query, args...)
		if err := row.Err(); err != nil {
			t.logger.WithError(err).Debug("failed to execute query")
		}
		return row
	}

	stmt, err := t.prepare(ctx, query)
	if err != nil {
		t.logger.WithError(err).WithField("query", query).Error("failed to prepare statement")
		return t.tx.QueryRowContext(ctx, query, args...)
	}

	row := stmt.QueryRowContext(ctx, args...)
	if err := row.Err(); err != nil {
		t.logger.WithError(err).Debug("failed to execute prepared statement")
	}
	return row
}

func (t *txConn) QueryxContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
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
