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
		return t.tx.QueryContext(ctx, query, args...)
	}

	stmt, err := t.prepare(ctx, query)
	if err != nil {
		t.logger.WithError(err).WithField("query", query).Error("failed to prepare statement")
		return t.tx.QueryContext(ctx, query, args...)
	}
	return stmt.QueryContext(ctx, args...)
}

func (t *txConn) QueryRowxContext(ctx context.Context, query string, args ...interface{}) *sqlx.Row {
	if !t.doPrepare {
		return t.tx.QueryRowxContext(ctx, query, args...)
	}

	stmt, err := t.prepare(ctx, query)
	if err != nil {
		t.logger.WithError(err).WithField("query", query).Error("failed to prepare statement")
		return t.tx.QueryRowxContext(ctx, query, args...)
	}
	return stmt.QueryRowxContext(ctx, args...)
}

func (t *txConn) QueryxContext(ctx context.Context, query string, args ...interface{}) (*sqlx.Rows, error) {
	if !t.doPrepare {
		return t.tx.QueryxContext(ctx, query, args...)
	}

	stmt, err := t.prepare(ctx, query)
	if err != nil {
		t.logger.WithError(err).WithField("query", query).Error("failed to prepare statement")
		return t.tx.QueryxContext(ctx, query, args...)
	}
	return stmt.QueryxContext(ctx, args...)
}

func (t *txConn) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	if !t.doPrepare {
		return t.tx.ExecContext(ctx, query, args...)
	}

	stmt, err := t.prepare(ctx, query)
	if err != nil {
		t.logger.WithError(err).WithField("query", query).Error("failed to prepare statement")
		return t.tx.ExecContext(ctx, query, args...)
	}
	return stmt.ExecContext(ctx, args...)
}
