package db

import (
	"context"
)

type MockHandle struct{}

var _ Handle = (*MockHandle)(nil)

func (h *MockHandle) WithTx(ctx context.Context, do func(ctx context.Context) error) (err error) {
	return do(ctx)
}

func (h *MockHandle) ReadOnly(ctx context.Context, do func(ctx context.Context) error) (err error) {
	return do(ctx)
}
