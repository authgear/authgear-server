package db

import (
	"context"
	"database/sql"
)

type stmtPreparer interface {
	PrepareContext(ctx context.Context, query string) (*sql.Stmt, error)
}

type txLike interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

type beginTxer interface {
	BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error)
}

// Handle allows a function to be run within a transaction.
type Handle interface {
	// WithTx runs do within a transaction.
	// If there is no error, the transaction is committed.
	WithTx(ctx context.Context, do func(ctx context.Context) error) (err error)

	// ReadOnly runs do within a transaction.
	// The transaction is always rolled back.
	ReadOnly(ctx context.Context, do func(ctx context.Context) error) (err error)
}

// PreparedStatementsHandle prepares and caches query.
type PreparedStatementsHandle interface {
	// WithTx runs do within a transaction.
	// If there is no error, the transaction is committed.
	WithTx(ctx context.Context, do func(ctx context.Context) error) (err error)
}
