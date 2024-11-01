package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/authgear/authgear-server/pkg/util/errorutil"
	"github.com/authgear/authgear-server/pkg/util/log"
	"github.com/authgear/authgear-server/pkg/util/uuid"
)

type HookHandle struct {
	Pool              *Pool
	ConnectionOptions ConnectionOptions
	Logger            *log.Logger

	tx    *txConn
	hooks []TransactionHook
}

func NewHookHandle(pool *Pool, opts ConnectionOptions, lf *log.Factory) *HookHandle {
	return &HookHandle{
		Pool:              pool,
		ConnectionOptions: opts,
		Logger:            lf.New("db-handle"),
	}
}

func (h *HookHandle) conn() (*txConn, error) {
	tx := h.tx
	if tx == nil {
		panic("hook-handle: transaction not started")
	}
	return tx, nil
}

func (h *HookHandle) UseHook(hook TransactionHook) {
	h.hooks = append(h.hooks, hook)
}

// WithTx commits if do finishes without error and rolls back otherwise.
func (h *HookHandle) WithTx(ctx context.Context, do func() error) (err error) {
	id := uuid.New()
	logger := h.Logger.WithField("debug_id", id)
	db, err := h.openDB()
	if err != nil {
		return
	}

	conn, err := db.db.Conn(ctx)
	if err != nil {
		err = fmt.Errorf("hook-handle: failed to acquire connection: %w", err)
		return
	}
	logger.Debug("acquire connection")

	tx, err := h.beginTx(ctx, logger, conn)
	if err != nil {
		return
	}

	// The assignment of h.tx can only be happen inside this method.
	// An invarant that must be held is that h.tx must be nil when this method terminates.
	// HookHandle can be used in multiple places.
	// Sometimes it is constructed per every request, and sometimes it is used throughout the entire lifetime of the process.
	// It is very important to make sure h.tx is nil after every call of WithTx or ReadOnly.
	// See https://github.com/authgear/authgear-server/issues/1612 for the bug of failing to enforcing the invariant.
	h.tx = tx
	defer func() {
		shouldRunDidCommitHooks := false

		// This defer block is run second.
		defer func() {
			// WillCommitTx of hook is allowed to access the database.
			// So the assignment to nil should happen last.
			h.tx = nil

			if shouldRunDidCommitHooks {
				// reset tx to complete the current transcation
				// before running the DidCommitTx hook
				// so new tx can be opened inside the DidCommitTx hook
				for _, hook := range h.hooks {
					hook.DidCommitTx()
				}
			}
		}()

		// This defer block is run first.
		defer func() {
			closeErr := conn.Close()
			if closeErr != nil && !errors.Is(closeErr, sql.ErrConnDone) {
				logger.WithError(closeErr).Error("failed to close connection")
			} else {
				logger.Debug("close connection")
			}
		}()

		if r := recover(); r != nil {
			_ = rollbackTx(tx)
			panic(r)
		} else if err != nil {
			_ = rollbackTx(tx)
		} else {
			err = commitTx(tx, h.hooks)
			if err == nil {
				shouldRunDidCommitHooks = true
			}
		}
	}()

	err = do()
	return
}

// ReadOnly runs do in a transaction and rolls back always.
func (h *HookHandle) ReadOnly(ctx context.Context, do func() error) (err error) {
	id := uuid.New()
	logger := h.Logger.WithField("debug_id", id)
	db, err := h.openDB()
	if err != nil {
		return
	}

	conn, err := db.db.Conn(ctx)
	if err != nil {
		err = fmt.Errorf("hook-handle: failed to acquire connection: %w", err)
		return
	}
	logger.Debug("acquire connection")

	tx, err := h.beginTx(ctx, logger, conn)
	if err != nil {
		return
	}

	// The assignment of h.tx can only be happen inside this method.
	// An invarant that must be held is that h.tx must be nil when this method terminates.
	// HookHandle can be used in multiple places.
	// Sometimes it is constructed per every request, and sometimes it is used throughout the entire lifetime of the process.
	// It is very important to make sure h.tx is nil after every call of WithTx or ReadOnly.
	// See https://github.com/authgear/authgear-server/issues/1612 for the bug of failing to enforcing the invariant.
	h.tx = tx
	defer func() {
		shouldRunDidCommitHooks := false

		defer func() {
			// WillCommitTx of hook is allowed to access the database.
			// So the assignment to nil should happen last.
			h.tx = nil

			if shouldRunDidCommitHooks {
				// reset tx to complete the current transcation
				// before running the DidCommitTx hook
				// so new tx can be opened inside the DidCommitTx hook
				for _, hook := range h.hooks {
					hook.DidCommitTx()
				}
			}
		}()

		// This defer block is run first.
		defer func() {
			closeErr := conn.Close()
			if closeErr != nil && !errors.Is(closeErr, sql.ErrConnDone) {
				logger.WithError(closeErr).Error("failed to close connection")
			} else {
				logger.Debug("close connection")
			}
		}()

		if r := recover(); r != nil {
			_ = rollbackTx(tx)
			panic(r)
		} else if err != nil {
			_ = rollbackTx(tx)
		} else {
			err = rollbackTx(tx)
			if err == nil {
				shouldRunDidCommitHooks = true
			}
		}
	}()

	err = do()
	return
}

func (h *HookHandle) beginTx(ctx context.Context, logger *log.Logger, conn *sql.Conn) (*txConn, error) {
	// Pass a nil TxOptions to use default isolation level.
	var txOptions *sql.TxOptions
	tx, err := conn.BeginTx(ctx, txOptions)
	if err != nil {
		return nil, fmt.Errorf("hook-handle: failed to begin transaction: %w", err)
	}

	logger.Debug("begin")

	return &txConn{
		conn:      conn,
		tx:        tx,
		logger:    logger,
		doPrepare: h.ConnectionOptions.UsePreparedStatements,
	}, nil
}

func commitTx(conn *txConn, hooks []TransactionHook) error {
	for _, hook := range hooks {
		err := hook.WillCommitTx()
		if err != nil {
			if rbErr := conn.tx.Rollback(); rbErr != nil {
				err = errorutil.WithSecondaryError(err, rbErr)
			}
			return err
		}
	}

	err := conn.tx.Commit()
	if err != nil {
		return fmt.Errorf("hook-handle: failed to commit transaction: %w", err)
	}
	conn.logger.Debug("commit")

	return nil
}

func rollbackTx(conn *txConn) error {
	err := conn.tx.Rollback()
	if err != nil {
		return fmt.Errorf("hook-handle: failed to rollback transaction: %w", err)
	}
	conn.logger.Debug("rollback")

	return nil
}

func (h *HookHandle) openDB() (*PoolDB, error) {
	h.Logger.WithFields(map[string]interface{}{
		"max_open_conns":             h.ConnectionOptions.MaxOpenConnection,
		"max_idle_conns":             h.ConnectionOptions.MaxIdleConnection,
		"conn_max_lifetime_seconds":  h.ConnectionOptions.MaxConnectionLifetime.Seconds(),
		"conn_max_idle_time_seconds": h.ConnectionOptions.IdleConnectionTimeout.Seconds(),
	}).Debug("open database")

	db, err := h.Pool.Open(h.ConnectionOptions)
	if err != nil {
		return nil, fmt.Errorf("hook-handle: failed to connect to database: %w", err)
	}

	return db, nil
}
