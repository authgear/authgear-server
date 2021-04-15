package global

import (
	"context"

	"github.com/authgear/authgear-server/pkg/lib/infra/db"
)

type SQLExecutor struct {
	db.SQLExecutor
}

func NewSQLExecutor(c context.Context, handle *Handle) *SQLExecutor {
	return &SQLExecutor{
		db.SQLExecutor{
			Context:  c,
			Database: handle,
		},
	}
}
