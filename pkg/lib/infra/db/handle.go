package db

import (
	"context"
)

// Handle allows a function to be run within a transaction.
type Handle interface {
	// WithTx runs do within a transaction.
	// If there is no error, the transaction is committed.
	WithTx(ctx context.Context, do func(ctx context.Context) error) (err error)

	// ReadOnly runs do within a transaction.
	// The transaction is always rolled back.
	ReadOnly(ctx context.Context, do func(ctx context.Context) error) (err error)

	// txConn allows internal access to the ongoing transaction.
	txConn(ctx context.Context) *txConn
}
