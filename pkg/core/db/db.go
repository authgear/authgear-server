package db

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"
)

type ExtContext interface {
	sqlx.ExtContext
	GetContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
}
