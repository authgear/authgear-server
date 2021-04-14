package db

import "github.com/jmoiron/sqlx"

type Handle interface {
	Conn() (sqlx.ExtContext, error)
	HasTx() bool
	WithTx(do func() error) (err error)
	ReadOnly(do func() error) (err error)
}
