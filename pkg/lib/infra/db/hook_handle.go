package db

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"

	"github.com/authgear/authgear-server/pkg/util/errorutil"
	"github.com/authgear/authgear-server/pkg/util/log"
)

type HookHandle struct {
	Context           context.Context
	Pool              *Pool
	ConnectionOptions ConnectionOptions
	Logger            *log.Logger

	tx    *sqlx.Tx
	hooks []TransactionHook
}

func NewHookHandle(ctx context.Context, pool *Pool, opts ConnectionOptions, lf *log.Factory) *HookHandle {
	return &HookHandle{
		Context:           ctx,
		Pool:              pool,
		ConnectionOptions: opts,
		Logger:            lf.New("db-handle"),
	}
}

func (h *HookHandle) conn() (sqlx.ExtContext, error) {
	tx := h.tx
	if tx == nil {
		panic("db: transaction not started")
	}
	return tx, nil
}

func (h *HookHandle) UseHook(hook TransactionHook) {
	h.hooks = append(h.hooks, hook)
}

// WithTx commits if do finishes without error and rolls back otherwise.
func (h *HookHandle) WithTx(do func() error) (err error) {
	if err = h.beginTx(); err != nil {
		return
	}

	defer func() {
		if r := recover(); r != nil {
			_ = h.rollbackTx()
			panic(r)
		} else if err != nil {
			_ = h.rollbackTx()
		} else {
			err = h.commitTx()
		}
	}()

	err = do()
	return
}

// ReadOnly runs do in a transaction and rolls back always.
func (h *HookHandle) ReadOnly(do func() error) (err error) {
	if err = h.beginTx(); err != nil {
		return
	}

	defer func() {
		if r := recover(); r != nil {
			_ = h.rollbackTx()
			panic(r)
		} else if err != nil {
			_ = h.rollbackTx()
		} else {
			err = h.rollbackTx()
		}
	}()

	err = do()
	return
}

func (h *HookHandle) beginTx() error {
	if h.tx != nil {
		panic("db: a transaction has already begun")
	}

	db, err := h.openDB()
	if err != nil {
		return err
	}
	tx, err := db.BeginTxx(h.Context, &sql.TxOptions{
		Isolation: sql.LevelRepeatableRead,
	})
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	h.tx = tx

	return nil
}

func (h *HookHandle) commitTx() error {
	if h.tx == nil {
		panic("db: a transaction has not begun")
	}

	for _, hook := range h.hooks {
		err := hook.WillCommitTx()
		if err != nil {
			if rbErr := h.tx.Rollback(); rbErr != nil {
				err = errorutil.WithSecondaryError(err, rbErr)
			}
			return err
		}
	}

	err := h.tx.Commit()
	if err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	h.tx = nil

	for _, hook := range h.hooks {
		hook.DidCommitTx()
	}

	return nil
}

func (h *HookHandle) rollbackTx() error {
	if h.tx == nil {
		panic("db: a transaction has not begun")
	}

	err := h.tx.Rollback()
	if err != nil {
		return fmt.Errorf("failed to rollback transaction: %w", err)
	}

	h.tx = nil
	return nil
}

func (h *HookHandle) openDB() (*sqlx.DB, error) {
	h.Logger.WithFields(map[string]interface{}{
		"max_open_conns":             h.ConnectionOptions.MaxOpenConnection,
		"max_idle_conns":             h.ConnectionOptions.MaxIdleConnection,
		"conn_max_lifetime_seconds":  h.ConnectionOptions.MaxConnectionLifetime.Seconds(),
		"conn_max_idle_time_seconds": h.ConnectionOptions.IdleConnectionTimeout.Seconds(),
	}).Debug("open database")

	db, err := h.Pool.Open(h.ConnectionOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	return db, nil
}
