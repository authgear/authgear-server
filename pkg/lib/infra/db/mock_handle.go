package db

import (
	"github.com/jmoiron/sqlx"
)

type MockHandle struct{}

func (h *MockHandle) conn() (sqlx.ExtContext, error) {
	panic("not mocked")
}

func (h *MockHandle) WithTx(do func() error) (err error) {
	return do()
}

func (h *MockHandle) ReadOnly(do func() error) (err error) {
	return do()
}
