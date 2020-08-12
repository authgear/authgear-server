package db

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"

	"github.com/authgear/authgear-server/pkg/auth/config"
	"github.com/authgear/authgear-server/pkg/util/errors"
	"github.com/authgear/authgear-server/pkg/util/log"
)

type Handle struct {
	ctx         context.Context
	pool        *Pool
	cfg         *config.DatabaseConfig
	credentials *config.DatabaseCredentials
	logger      *log.Logger

	db    *sqlx.DB
	tx    *sqlx.Tx
	hooks []TransactionHook
}

func NewHandle(ctx context.Context, pool *Pool, cfg *config.DatabaseConfig, credentials *config.DatabaseCredentials, lf *log.Factory) *Handle {
	return &Handle{
		ctx:         ctx,
		pool:        pool,
		cfg:         cfg,
		credentials: credentials,
		logger:      lf.New("db-handle"),
	}
}

func (h *Handle) Conn() (sqlx.ExtContext, error) {
	tx := h.tx
	if tx == nil {
		return h.openDB()
	}
	return tx, nil
}

func (h *Handle) HasTx() bool {
	return h.tx != nil
}

func (h *Handle) UseHook(hook TransactionHook) {
	h.hooks = append(h.hooks, hook)
}

// WithTx commits if do finishes without error and rolls back otherwise.
func (h *Handle) WithTx(do func() error) (err error) {
	if err = h.beginTx(); err != nil {
		return
	}

	defer func() {
		if r := recover(); r != nil {
			_ = h.rollbackTx()
			panic(r)
		} else if err != nil {
			if rbErr := h.rollbackTx(); rbErr != nil {
				h.logger.WithError(rbErr).Error("failed to rollback tx")
			}
		} else {
			err = h.commitTx()
		}
	}()

	err = do()
	return
}

// ReadOnly runs do in a transaction and rolls back always.
func (h *Handle) ReadOnly(do func() error) (err error) {
	if err = h.beginTx(); err != nil {
		return
	}

	defer func() {
		if r := recover(); r != nil {
			_ = h.rollbackTx()
			panic(r)
		} else if err != nil {
			if rbErr := h.rollbackTx(); rbErr != nil {
				h.logger.WithError(rbErr).Error("failed to rollback tx")
			}
		} else {
			err = h.rollbackTx()
		}
	}()

	err = do()
	return
}

func (h *Handle) beginTx() error {
	if h.tx != nil {
		panic("db: a transaction has already begun")
	}

	db, err := h.openDB()
	if err != nil {
		return err
	}
	tx, err := db.BeginTxx(h.ctx, &sql.TxOptions{
		Isolation: sql.LevelRepeatableRead,
	})
	if err != nil {
		return errors.HandledWithMessage(err, "failed to begin transaction")
	}

	h.tx = tx

	return nil
}

func (h *Handle) commitTx() error {
	if h.tx == nil {
		panic("db: a transaction has not begun")
	}

	for _, hook := range h.hooks {
		err := hook.WillCommitTx()
		if err != nil {
			if rbErr := h.tx.Rollback(); rbErr != nil {
				err = errors.WithSecondaryError(err, rbErr)
			}
			return err
		}
	}

	err := h.tx.Commit()
	if err != nil {
		return errors.HandledWithMessage(err, "failed to commit transaction")
	}
	h.tx = nil

	for _, hook := range h.hooks {
		hook.DidCommitTx()
	}

	return nil
}

func (h *Handle) rollbackTx() error {
	if h.tx == nil {
		panic("db: a transaction has not begun")
	}

	err := h.tx.Rollback()
	if err != nil {
		return errors.HandledWithMessage(err, "failed to rollback transaction")
	}

	h.tx = nil
	return nil
}

func (h *Handle) openDB() (*sqlx.DB, error) {
	if h.db == nil {
		opts := OpenOptions{
			URL:             h.credentials.DatabaseURL,
			MaxOpenConns:    *h.cfg.MaxOpenConnection,
			MaxIdleConns:    *h.cfg.MaxIdleConnection,
			ConnMaxLifetime: h.cfg.MaxConnectionLifetime.Duration(),
		}
		h.logger.WithFields(map[string]interface{}{
			"max_open_conns":            opts.MaxOpenConns,
			"max_idle_conns":            opts.MaxIdleConns,
			"conn_max_lifetime_seconds": opts.ConnMaxLifetime.Seconds(),
		}).Debug("open database")

		db, err := h.pool.Open(opts)
		if err != nil {
			return nil, errors.HandledWithMessage(err, "failed to connect to database")
		}

		h.db = db
	}

	return h.db, nil
}
