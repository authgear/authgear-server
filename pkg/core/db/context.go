package db

import (
	"context"
	"database/sql"
	"net/http"

	"github.com/jmoiron/sqlx"

	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/errors"
)

type contextKey string

var (
	keyContainer = contextKey("container")
)

// Context provides db with the interface for retrieving an interface to execute sql
type Context interface {
	DB() (ExtContext, error)
}

// TxContext provides the interface for managing transaction
type TxContext interface {
	beginTx() error
	commitTx() error
	rollbackTx() error

	HasTx() bool
	UseHook(TransactionHook)
}

// WithTx commits if do finishes without error and rolls back otherwise.
func WithTx(tx TxContext, do func() error) (err error) {
	if err = tx.beginTx(); err != nil {
		return
	}

	defer func() {
		if r := recover(); r != nil {
			_ = tx.rollbackTx()
			panic(r)
		} else if err != nil {
			if rbErr := tx.rollbackTx(); rbErr != nil {
				err = errors.WithSecondaryError(err, rbErr)
			}
		} else {
			err = tx.commitTx()
		}
	}()

	err = do()
	return
}

// ReadOnly runs do in a transaction and rolls back always.
func ReadOnly(tx TxContext, do func() error) (err error) {
	if err = tx.beginTx(); err != nil {
		return
	}

	defer func() {
		if r := recover(); r != nil {
			_ = tx.rollbackTx()
			panic(r)
		} else if err != nil {
			if rbErr := tx.rollbackTx(); rbErr != nil {
				err = errors.WithSecondaryError(err, rbErr)
			}
		} else {
			err = tx.rollbackTx()
		}
	}()

	err = do()
	return
}

type contextContainer struct {
	pool  Pool
	db    *sqlx.DB
	tx    *sqlx.Tx
	hooks []TransactionHook
}

type dbContext struct {
	context.Context
	tConfig config.TenantConfiguration
}

func InitDBContext(ctx context.Context, pool Pool) context.Context {
	container := &contextContainer{pool: pool}
	return context.WithValue(ctx, keyContainer, container)
}

// InitRequestDBContext initialize db context for the request
func InitRequestDBContext(req *http.Request, pool Pool) *http.Request {
	return req.WithContext(InitDBContext(req.Context(), pool))
}

func newDBContext(ctx context.Context, tConfig config.TenantConfiguration) *dbContext {
	return &dbContext{Context: ctx, tConfig: tConfig}
}

// NewContextWithContext creates a new context.DB from context
func NewContextWithContext(ctx context.Context, tConfig config.TenantConfiguration) Context {
	return newDBContext(ctx, tConfig)
}

// NewTxContextWithContext creates a new context.Tx from context
func NewTxContextWithContext(ctx context.Context, tConfig config.TenantConfiguration) TxContext {
	return newDBContext(ctx, tConfig)
}

func (d *dbContext) DB() (ExtContext, error) {
	if d.tx() != nil {
		return d.tx(), nil
	}

	return d.lazydb()
}

func (d *dbContext) HasTx() bool {
	return d.tx() != nil
}

func (d *dbContext) UseHook(h TransactionHook) {
	container := d.container()
	container.hooks = append(container.hooks, h)
}

func (d *dbContext) beginTx() error {
	if d.tx() != nil {
		panic("skydb: a transaction has already begun")
	}

	db, err := d.lazydb()
	if err != nil {
		return err
	}
	tx, err := db.BeginTxx(d, &sql.TxOptions{
		Isolation: sql.LevelRepeatableRead,
	})
	if err != nil {
		return errors.HandledWithMessage(err, "failed to begin transaction")
	}

	container := d.container()
	container.tx = tx

	return nil
}

func (d *dbContext) commitTx() error {
	if d.tx() == nil {
		panic("skydb: a transaction has not begun")
	}

	container := d.container()
	for _, hook := range container.hooks {
		err := hook.WillCommitTx()
		if err != nil {
			if rbErr := container.tx.Rollback(); rbErr != nil {
				err = errors.WithSecondaryError(err, rbErr)
			}
			return err
		}
	}

	err := container.tx.Commit()
	if err != nil {
		return errors.HandledWithMessage(err, "failed to commit transaction")
	}
	container.tx = nil

	for _, hook := range container.hooks {
		hook.DidCommitTx()
	}

	return nil
}

func (d *dbContext) rollbackTx() error {
	if d.tx() == nil {
		panic("skydb: a transaction has not begun")
	}

	err := d.tx().Rollback()
	if err != nil {
		return errors.HandledWithMessage(err, "failed to rollback transaction")
	}

	container := d.container()
	container.tx = nil
	return nil
}

func (d *dbContext) db() *sqlx.DB {
	return d.container().db
}

func (d *dbContext) tx() *sqlx.Tx {
	return d.container().tx
}

func (d *dbContext) lazydb() (*sqlx.DB, error) {
	db := d.db()
	if db == nil {
		container := d.container()

		var err error
		if db, err = container.pool.Open(d.tConfig); err != nil {
			return nil, errors.HandledWithMessage(err, "failed to connect to database")
		}

		container.db = db
	}

	return db, nil
}

func (d *dbContext) container() *contextContainer {
	return d.Value(keyContainer).(*contextContainer)
}
