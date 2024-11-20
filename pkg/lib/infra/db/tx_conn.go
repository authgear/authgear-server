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
	// Let me try to explain what is happening here.
	// First, I need to let you know a few facts.
	//
	// 1. *Stmt produced by *DB can be reused in different *Conn.
	// 2. *DB.PrepareContext acquires an extra connection behind the scene, thus it will cause deadlock.
	// 3. *Stmt produced by *Conn is bound to the underlying driver.Conn.
	// 4. *Stmt prodcued by *Tx.PrepareContext or *Tx.StmtContext is bound the lifetime of *Tx. This means
	//    when the *Tx rolls back or commits, the *Stmt is closed.
	// 5. *Tx.StmtContext always re-prepares a new *Stmt, if the *Stmt passed to it is produced by a *Conn.
	//
	// With these facts, we have the following conclusions.
	// 6. Given 2, we cannot use *DB.PrepareContext. We can only use *Conn.PrepareContext or *Tx.PrepareContext.
	// 7. Given 4 and 5, using transaction-bound *Stmt does not benefit a lot, as the prepared statement is created and closed shortly.
	// 8. So we can only use *Conn.PrepareContext.
	//
	// How can we close the *Stmt produced by *Conn.PrepareContext?
	// The answer is we cannot, because *Conn.Close means returning the underlying driver.Conn to the pool.
	// A returned driver.Conn is still alive, so are the driver.Stmt it has prepared so far.
	//
	// Therefore, this function makes an assumption that the underlying driver.Conn
	// will take care of closing the driver.Stmt when it is closed.
	stmt, err := t.conn.PrepareContext(ctx, query)
	if err != nil {
		return nil, err
	}

	return stmt, nil
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
