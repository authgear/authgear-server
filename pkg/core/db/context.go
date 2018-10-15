package db

import (
	"context"
	"net/http"

	"github.com/jmoiron/sqlx"
)

type contextKey string

var (
	keyContainer = contextKey("container")
)

// DB provides the interface for retrieving an interface to execute sql
type Context interface {
	DB() ExtContext
}

// Tx provides the interface for managing transaction
type TxContext interface {
	HasTx() bool
	BeginTx() error
	CommitTx() error
	RollbackTx() error
}

// TODO: handle thread safety
type contextContainer struct {
	db *sqlx.DB
	tx *sqlx.Tx
}

type dbContext struct {
	context.Context
	dbOpener func() (*sqlx.DB, error)
}

// InitRequestDBContext initialize db context for the request
func InitRequestDBContext(req *http.Request) *http.Request {
	container := &contextContainer{}
	return req.WithContext(context.WithValue(req.Context(), keyContainer, container))
}

// NewContextWithContext creates a new context.DB from context
func NewContextWithContext(ctx context.Context, dbOpener func() (*sqlx.DB, error)) Context {
	return &dbContext{
		Context:  ctx,
		dbOpener: dbOpener,
	}
}

// NewTxContextWithContext creates a new context.Tx from context
func NewTxContextWithContext(ctx context.Context, dbOpener func() (*sqlx.DB, error)) TxContext {
	return &dbContext{
		Context:  ctx,
		dbOpener: dbOpener,
	}
}

func (d *dbContext) DB() ExtContext {
	if d.tx() != nil {
		return d.tx()
	}

	return d.lazydb()
}

func (d *dbContext) HasTx() bool {
	return d.tx() != nil
}

func (d *dbContext) BeginTx() (err error) {
	if d.tx() != nil {
		err = ErrDatabaseTxDidBegin
		return err
	}

	if tx, err := d.lazydb().BeginTxx(d, nil); err == nil {
		container := d.container()
		container.tx = tx
	}

	return
}

func (d *dbContext) CommitTx() (err error) {
	if d.tx() == nil {
		err = ErrDatabaseTxDidNotBegin
		return err
	}

	if err = d.tx().Commit(); err == nil {
		container := d.container()
		container.tx = nil
	}

	return
}

func (d *dbContext) RollbackTx() (err error) {
	if d.tx() == nil {
		err = ErrDatabaseTxDidNotBegin
		return
	}

	if err = d.tx().Rollback(); err == nil {
		container := d.container()
		container.tx = nil
	}

	return
}

func (d *dbContext) db() *sqlx.DB {
	return d.container().db
}

func (d *dbContext) tx() *sqlx.Tx {
	return d.container().tx
}

func (d *dbContext) lazydb() *sqlx.DB {
	db := d.db()
	if db == nil {
		var err error
		if db, err = d.dbOpener(); err != nil {
			panic(err)
		}

		container := d.container()
		container.db = db
	}

	return db
}

func (d *dbContext) container() *contextContainer {
	return d.Value(keyContainer).(*contextContainer)
}
