package db

import (
	"context"
)

type MockHandle struct{}

func (h *MockHandle) conn() (*txConn, error) {
	panic("not mocked")
}

func (h *MockHandle) WithTx(ctx context.Context, do func(ctx context.Context) error) (err error) {
	return do(ctx)
}

func (h *MockHandle) ReadOnly(ctx context.Context, do func(ctx context.Context) error) (err error) {
	return do(ctx)
}
