package db

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"

	"github.com/authgear/authgear-server/pkg/util/errorutil"
	"github.com/authgear/authgear-server/pkg/util/otelutil/oteldatabasesql"
	"github.com/authgear/authgear-server/pkg/util/slogutil"
	"github.com/authgear/authgear-server/pkg/util/uuid"
)

type hooksContextKeyType struct{}

var hooksContextKey = hooksContextKeyType{}

type hooksContextValue struct {
	Hooks []TransactionHook
}

func contextWithHooks(ctx context.Context, value *hooksContextValue) context.Context {
	return context.WithValue(ctx, hooksContextKey, value)
}

func contextGetHooks(ctx context.Context) (*hooksContextValue, bool) {
	v, ok := ctx.Value(hooksContextKey).(*hooksContextValue)
	if !ok {
		return nil, false
	}
	return v, true
}

func mustContextGetHooks(ctx context.Context) *hooksContextValue {
	v, ok := contextGetHooks(ctx)
	if !ok {
		panic(fmt.Errorf("programming_error: hooks is not initialized"))
	}
	return v
}

type txLikeContextKeyType struct{}

var txLikeContextKey = txLikeContextKeyType{}

type txLikeContextValue struct {
	TxLike txLike
}

func contextWithTxLike(ctx context.Context, value *txLikeContextValue) context.Context {
	return context.WithValue(ctx, txLikeContextKey, value)
}

func contextGetTxLike(ctx context.Context) (*txLikeContextValue, bool) {
	v, ok := ctx.Value(txLikeContextKey).(*txLikeContextValue)
	if !ok {
		return nil, false
	}
	return v, true
}

func mustContextGetTxLike(ctx context.Context) *txLikeContextValue {
	v, ok := contextGetTxLike(ctx)
	if !ok {
		panic(fmt.Errorf("programming_error: tx is not initialized"))
	}
	return v
}

type HookHandle struct {
	Pool              Pool_
	ConnectionInfo    ConnectionInfo
	ConnectionOptions ConnectionOptions
}

func mustGetTxLike(ctx context.Context) txLike {
	return mustContextGetTxLike(ctx).TxLike
}

var _ Handle = (*HookHandle)(nil)

var HookHandleLogger = slogutil.NewLogger("db-handle")

func NewHookHandle(pool Pool_, info ConnectionInfo, opts ConnectionOptions) *HookHandle {
	return &HookHandle{
		Pool:              pool,
		ConnectionInfo:    info,
		ConnectionOptions: opts,
	}
}

func (h *HookHandle) UseHook(ctx context.Context, hook TransactionHook) {
	v := mustContextGetHooks(ctx)
	v.Hooks = append(v.Hooks, hook)
}

// WithTx commits if do finishes without error and rolls back otherwise.
// WithTx is reentrant, meaning that you can call WithTx even when a previous WithTx does not finish yet.
// Normally you should not call WithTx within a WithTx, but there is a legit use case.
//
//	// Assume ctx is a http.Request context.
//	h.WithTx(ctx, func(ctx context.Context) error {
//		// ctx here is associated with a *sql.Tx (Tx1)
//		go func() {
//			// ctx is detached from the http.Request context.
//			ctx = ctx.WithCancel(ctx)
//			h.WithTx(ctx, func(ctx context.Context) error {
//				// ctx is associated with a *sqlTx (Tx2)
//			})
//		}()
//	})
func (h *HookHandle) WithTx(ctx_original context.Context, do func(ctx context.Context) error) (err error) {
	ctx_hooks := contextWithHooks(ctx_original, &hooksContextValue{})
	shouldRunDidCommitHooks := false
	defer func() {
		if shouldRunDidCommitHooks {
			for _, hook := range mustContextGetHooks(ctx_hooks).Hooks {
				hook.DidCommitTx(ctx_hooks)
			}
		}
	}()

	id := uuid.New()
	logger := slogutil.GetContextLogger(ctx_hooks)
	logger = logger.With(slog.String("debug_id", id))
	ctx_hooks = slogutil.SetContextLogger(ctx_hooks, logger)

	db, err := h.openDB(ctx_hooks)
	if err != nil {
		return
	}

	conn, err := db.Conn(ctx_hooks)
	if err != nil {
		return
	}
	defer func() {
		_ = conn.Close()
	}()

	tx, err := beginTx(ctx_hooks, conn)
	if err != nil {
		return
	}

	ctx_hooks_tx := contextWithTxLike(ctx_hooks, &txLikeContextValue{
		TxLike: tx,
	})

	defer func() {
		if r := recover(); r != nil {
			_ = rollbackTx(ctx_hooks_tx, tx)
			panic(r)
		} else if err != nil {
			_ = rollbackTx(ctx_hooks_tx, tx)
		} else {
			err = commitTx(ctx_hooks_tx, tx, mustContextGetHooks(ctx_hooks_tx).Hooks)
			if err == nil {
				shouldRunDidCommitHooks = true
			}
		}
	}()

	err = do(ctx_hooks_tx)
	return
}

// ReadOnly is like WithTx, except that it always rolls back.
func (h *HookHandle) ReadOnly(ctx_original context.Context, do func(ctx context.Context) error) (err error) {
	ctx_hooks := contextWithHooks(ctx_original, &hooksContextValue{})
	shouldRunDidCommitHooks := false
	defer func() {
		if shouldRunDidCommitHooks {
			for _, hook := range mustContextGetHooks(ctx_hooks).Hooks {
				hook.DidCommitTx(ctx_hooks)
			}
		}
	}()

	id := uuid.New()
	logger := slogutil.GetContextLogger(ctx_hooks)
	logger = logger.With(slog.String("debug_id", id))
	ctx_hooks = slogutil.SetContextLogger(ctx_hooks, logger)

	db, err := h.openDB(ctx_hooks)
	if err != nil {
		return
	}

	conn, err := db.Conn(ctx_hooks)
	if err != nil {
		return
	}
	defer func() {
		_ = conn.Close()
	}()

	tx, err := beginTx(ctx_hooks, conn)
	if err != nil {
		return
	}

	ctx_hooks_tx := contextWithTxLike(ctx_hooks, &txLikeContextValue{
		TxLike: tx,
	})

	defer func() {
		if r := recover(); r != nil {
			_ = rollbackTx(ctx_hooks_tx, tx)
			panic(r)
		} else if err != nil {
			_ = rollbackTx(ctx_hooks_tx, tx)
		} else {
			err = rollbackTx(ctx_hooks_tx, tx)
			if err == nil {
				shouldRunDidCommitHooks = true
			}
		}
	}()

	err = do(ctx_hooks_tx)
	return
}

func (h *HookHandle) WithPrepareStatementsHandle(ctx context.Context, do func(ctx context.Context, handle PreparedStatementsHandle) error) (err error) {
	id := uuid.New()
	{
		logger := slogutil.GetContextLogger(ctx)
		logger = logger.With(slog.String("debug_id", id))
		ctx = slogutil.SetContextLogger(ctx, logger)
	}
	logger := HookHandleLogger.GetLogger(ctx)

	db, err := h.openDB(ctx)
	if err != nil {
		return
	}

	conn, err := db.Conn(ctx)
	if err != nil {
		err = fmt.Errorf("hook-handle: failed to acquire connection: %w", err)
		return
	}
	logger.Debug(ctx, "acquire connection")

	preparedStatementsHandle := &preparedStatementsHandle{
		conn:             conn,
		cachedStatements: make(map[string]*sql.Stmt),
	}
	defer func() {
		_ = preparedStatementsHandle.Close(ctx)
	}()

	ctx = withPreparedStatementsHandle(ctx, preparedStatementsHandle)

	err = do(ctx, preparedStatementsHandle)
	return
}

func (*HookHandle) IsInTx(ctx context.Context) bool {
	_, isInTx := contextGetTxLike(ctx)
	return isInTx
}

func beginTx(ctx context.Context, conn oteldatabasesql.Conn_) (*sql.Tx, error) {
	logger := HookHandleLogger.GetLogger(ctx)

	// Pass a nil TxOptions to use default isolation level.
	var txOptions *sql.TxOptions
	tx, err := conn.BeginTx(ctx, txOptions)
	if err != nil {
		return nil, fmt.Errorf("hook-handle: failed to begin transaction: %w", err)
	}

	logger.Debug(ctx, "begin")
	return tx, nil
}

func commitTx(ctx context.Context, tx *sql.Tx, hooks []TransactionHook) error {
	logger := HookHandleLogger.GetLogger(ctx)

	for _, hook := range hooks {
		err := hook.WillCommitTx(ctx)
		if err != nil {
			if rbErr := tx.Rollback(); rbErr != nil {
				err = errorutil.WithSecondaryError(err, rbErr)
			}
			return err
		}
	}

	err := tx.Commit()
	if err != nil {
		return fmt.Errorf("hook-handle: failed to commit transaction: %w", err)
	}
	logger.Debug(ctx, "commit")
	return nil
}

func rollbackTx(ctx context.Context, tx *sql.Tx) error {
	logger := HookHandleLogger.GetLogger(ctx)

	err := tx.Rollback()
	if err != nil {
		return fmt.Errorf("hook-handle: failed to rollback transaction: %w", err)
	}
	logger.Debug(ctx, "rollback")

	return nil
}

func (h *HookHandle) openDB(ctx context.Context) (oteldatabasesql.ConnPool_, error) {
	logger := HookHandleLogger.GetLogger(ctx)
	logger.Debug(ctx, "open database",
		slog.String("purpose", string(h.ConnectionInfo.Purpose)),
		slog.Int("max_open_conns", h.ConnectionOptions.MaxOpenConnection),
		slog.Int("max_idle_conns", h.ConnectionOptions.MaxIdleConnection),
		slog.Duration("conn_max_lifetime_seconds", h.ConnectionOptions.MaxConnectionLifetime),
		slog.Duration("conn_max_idle_time_seconds", h.ConnectionOptions.IdleConnectionTimeout),
	)

	db, err := h.Pool.Open(h.ConnectionInfo, h.ConnectionOptions)
	if err != nil {
		return nil, fmt.Errorf("hook-handle: failed to connect to database: %w", err)
	}

	return db, nil
}
