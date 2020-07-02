package db

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"

	"github.com/authgear/authgear-server/pkg/auth/config"
	"github.com/authgear/authgear-server/pkg/core/errors"
	"github.com/authgear/authgear-server/pkg/log"
)

type ExtContext = sqlx.ExtContext

type Context interface {
	DB() (ExtContext, error)
	HasTx() bool
	UseHook(TransactionHook)
	WithTx(do func() error) error
	ReadOnly(do func() error) error
}

type dbContext struct {
	context.Context
	pool        *Pool
	cfg         *config.DatabaseConfig
	credentials *config.DatabaseCredentials
	logger      *log.Logger

	db    *sqlx.DB
	tx    *sqlx.Tx
	hooks []TransactionHook
}

func NewContext(ctx context.Context, pool *Pool, cfg *config.DatabaseConfig, credentials *config.DatabaseCredentials, lf *log.Factory) Context {
	return &dbContext{
		Context:     ctx,
		pool:        pool,
		cfg:         cfg,
		credentials: credentials,
		logger:      lf.New("dbcontext"),
	}
}

func (ctx *dbContext) DB() (ExtContext, error) {
	tx := ctx.tx
	if tx == nil {
		return ctx.openDB()
	}
	return tx, nil
}

func (ctx *dbContext) HasTx() bool {
	return ctx.tx != nil
}

func (ctx *dbContext) UseHook(h TransactionHook) {
	ctx.hooks = append(ctx.hooks, h)
}

// WithTx commits if do finishes without error and rolls back otherwise.
func (ctx *dbContext) WithTx(do func() error) (err error) {
	if err = ctx.beginTx(); err != nil {
		return
	}

	defer func() {
		if r := recover(); r != nil {
			_ = ctx.rollbackTx()
			panic(r)
		} else if err != nil {
			if rbErr := ctx.rollbackTx(); rbErr != nil {
				ctx.logger.WithError(rbErr).Error("failed to rollback tx")
			}
		} else {
			err = ctx.commitTx()
		}
	}()

	err = do()
	return
}

// ReadOnly runs do in a transaction and rolls back always.
func (ctx *dbContext) ReadOnly(do func() error) (err error) {
	if err = ctx.beginTx(); err != nil {
		return
	}

	defer func() {
		if r := recover(); r != nil {
			_ = ctx.rollbackTx()
			panic(r)
		} else if err != nil {
			if rbErr := ctx.rollbackTx(); rbErr != nil {
				ctx.logger.WithError(rbErr).Error("failed to rollback tx")
			}
		} else {
			err = ctx.rollbackTx()
		}
	}()

	err = do()
	return
}

func (ctx *dbContext) beginTx() error {
	if ctx.tx != nil {
		panic("db: a transaction has already begun")
	}

	db, err := ctx.openDB()
	if err != nil {
		return err
	}
	tx, err := db.BeginTxx(ctx.Context, &sql.TxOptions{
		Isolation: sql.LevelRepeatableRead,
	})
	if err != nil {
		return errors.HandledWithMessage(err, "failed to begin transaction")
	}

	ctx.tx = tx

	return nil
}

func (ctx *dbContext) commitTx() error {
	if ctx.tx == nil {
		panic("db: a transaction has not begun")
	}

	for _, hook := range ctx.hooks {
		err := hook.WillCommitTx()
		if err != nil {
			if rbErr := ctx.tx.Rollback(); rbErr != nil {
				err = errors.WithSecondaryError(err, rbErr)
			}
			return err
		}
	}

	err := ctx.tx.Commit()
	if err != nil {
		return errors.HandledWithMessage(err, "failed to commit transaction")
	}
	ctx.tx = nil

	for _, hook := range ctx.hooks {
		hook.DidCommitTx()
	}

	return nil
}

func (ctx *dbContext) rollbackTx() error {
	if ctx.tx == nil {
		panic("db: a transaction has not begun")
	}

	err := ctx.tx.Rollback()
	if err != nil {
		return errors.HandledWithMessage(err, "failed to rollback transaction")
	}

	ctx.tx = nil
	return nil
}

func (ctx *dbContext) openDB() (*sqlx.DB, error) {
	if ctx.db == nil {
		opts := OpenOptions{
			URL:             ctx.credentials.DatabaseURL,
			MaxOpenConns:    *ctx.cfg.MaxOpenConnection,
			MaxIdleConns:    *ctx.cfg.MaxIdleConnection,
			ConnMaxLifetime: ctx.cfg.MaxConnectionLifetime.Duration(),
		}
		ctx.logger.WithFields(map[string]interface{}{
			"max_open_conns":            opts.MaxOpenConns,
			"max_idle_conns":            opts.MaxIdleConns,
			"conn_max_lifetime_seconds": opts.ConnMaxLifetime.Seconds(),
		}).Debug("open database")

		db, err := ctx.pool.Open(opts)
		if err != nil {
			return nil, errors.HandledWithMessage(err, "failed to connect to database")
		}

		ctx.db = db
	}

	return ctx.db, nil
}
