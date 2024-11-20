package db

import (
	"context"
	"database/sql"
)

type ConnLike interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

// Handle allows a function to be run within a transaction.
type Handle interface {
	// WithTx runs do within a transaction.
	// If there is no error, the transaction is committed.
	WithTx(ctx context.Context, do func(ctx context.Context) error) (err error)

	// ReadOnly runs do within a transaction.
	// The transaction is always rolled back.
	ReadOnly(ctx context.Context, do func(ctx context.Context) error) (err error)

	// connLike allows internal access to the ongoing transaction.
	connLike(ctx context.Context) ConnLike
}
