package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"

	"github.com/authgear/authgear-server/pkg/util/otelutil/oteldatabasesql"
	"github.com/authgear/authgear-server/pkg/util/slogutil"
)

var preparedStatementsHandleLogger = slogutil.NewLogger("prepared-statements-handle")

type preparedStatementsHandleContextKeyType struct{}

var preparedStatementsHandleContextKey = preparedStatementsHandleContextKeyType{}

type preparedStatementsHandle struct {
	conn             oteldatabasesql.Conn_
	cachedStatements map[string]*sql.Stmt
}

func withPreparedStatementsHandle(ctx context.Context, value *preparedStatementsHandle) context.Context {
	return context.WithValue(ctx, preparedStatementsHandleContextKey, value)
}

func getPreparedStatementsHandle(ctx context.Context) (*preparedStatementsHandle, bool) {
	v, ok := ctx.Value(preparedStatementsHandleContextKey).(*preparedStatementsHandle)
	if !ok {
		return nil, false
	}
	return v, true
}

func getStmtPreparer(ctx context.Context) (stmtPreparer, bool) {
	h, ok := getPreparedStatementsHandle(ctx)
	if !ok {
		return nil, false
	}

	return h, true
}

func (h *preparedStatementsHandle) PrepareContext(ctx context.Context, query string) (*sql.Stmt, error) {
	logger := preparedStatementsHandleLogger.GetLogger(ctx)

	stmt, ok := h.cachedStatements[query]
	if ok {
		logger.Debug(ctx, "prepared statement cache hit",
			slog.String("conn", fmt.Sprintf("%p", h.conn)),
			slog.String("stmt", fmt.Sprintf("%p", stmt)),
		)
		return stmt, nil
	}

	stmt, err := h.conn.PrepareContext(ctx, query)
	if err != nil {
		return nil, err
	}
	h.cachedStatements[query] = stmt
	logger.Debug(ctx, "prepared statement cache miss",
		slog.String("conn", fmt.Sprintf("%p", h.conn)),
		slog.String("stmt", fmt.Sprintf("%p", stmt)),
	)
	return stmt, nil
}

func (h *preparedStatementsHandle) Close(ctx context.Context) error {
	logger := preparedStatementsHandleLogger.GetLogger(ctx)
	logger.Debug(ctx, "start closing prepared statement handle",
		slog.String("conn", fmt.Sprintf("%p", h.conn)),
	)

	var err error
	for _, stmt := range h.cachedStatements {
		closeErr := stmt.Close()
		err = errors.Join(err, closeErr)
	}

	closeErr := h.conn.Close()
	err = errors.Join(err, closeErr)

	logger.WithError(err).Debug(ctx, "end closing prepared statement handle",
		slog.String("conn", fmt.Sprintf("%p", h.conn)),
	)

	return err
}

func (h *preparedStatementsHandle) WithTx(ctx_original context.Context, do func(ctx context.Context) error) (err error) {
	ctx_hooks := contextWithHooks(ctx_original, &hooksContextValue{})
	shouldRunDidCommitHooks := false
	defer func() {
		if shouldRunDidCommitHooks {
			for _, hook := range mustContextGetHooks(ctx_hooks).Hooks {
				hook.DidCommitTx(ctx_hooks)
			}
		}
	}()

	err = beginTx(ctx_hooks, h.conn, func(tx *sql.Tx) (err error) {
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
		return err
	})
	return
}
