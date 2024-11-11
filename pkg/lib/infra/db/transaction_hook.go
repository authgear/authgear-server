package db

import (
	"context"
)

type TransactionHook interface {
	WillCommitTx(ctx context.Context) error
	DidCommitTx(ctx context.Context)
}
