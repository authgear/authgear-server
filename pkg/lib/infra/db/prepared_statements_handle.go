package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/authgear/authgear-server/pkg/util/log"
)

type preparedStatementsHandleContextKeyType struct{}

var preparedStatementsHandleContextKey = preparedStatementsHandleContextKeyType{}

type preparedStatementsHandle struct {
	logger           *log.Logger
	conn             *sql.Conn
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
	stmt, ok := h.cachedStatements[query]
	if ok {
		h.logger.
			WithField("conn", fmt.Sprintf("%p", h.conn)).
			WithField("stmt", fmt.Sprintf("%p", stmt)).
			Debug("prepared statement cache hit")
		return stmt, nil
	}

	stmt, err := h.conn.PrepareContext(ctx, query)
	if err != nil {
		return nil, err
	}
	h.cachedStatements[query] = stmt
	h.logger.
		WithField("conn", fmt.Sprintf("%p", h.conn)).
		WithField("stmt", fmt.Sprintf("%p", stmt)).
		Debug("prepared statement cache miss")
	return stmt, nil
}

func (h *preparedStatementsHandle) Close() error {
	h.logger.
		WithField("conn", fmt.Sprintf("%p", h.conn)).
		Debug("start closing prepared statement handle")

	var err error
	for _, stmt := range h.cachedStatements {
		closeErr := stmt.Close()
		err = errors.Join(err, closeErr)
	}

	closeErr := h.conn.Close()
	err = errors.Join(err, closeErr)

	h.logger.
		WithError(err).
		WithField("conn", fmt.Sprintf("%p", h.conn)).
		Debug("end closing prepared statement handle")

	return err
}

func (h *preparedStatementsHandle) WithTx(ctx context.Context, do func(ctx context.Context) error) (err error) {
	tx, err := beginTx(ctx, h.logger, h.conn)
	if err != nil {
		return
	}

	ctx = hookHandleContextWithValue(ctx, &hookHandleContextValue{
		TxLike: tx,
	})

	shouldRunDidCommitHooks := false

	defer func() {
		if shouldRunDidCommitHooks {
			for _, hook := range mustHookHandleContextGetValue(ctx).Hooks {
				hook.DidCommitTx(ctx)
			}
		}
	}()

	defer func() {
		if r := recover(); r != nil {
			_ = rollbackTx(h.logger, tx)
			panic(r)
		} else if err != nil {
			_ = rollbackTx(h.logger, tx)
		} else {
			err = commitTx(ctx, h.logger, tx, mustHookHandleContextGetValue(ctx).Hooks)
			if err == nil {
				shouldRunDidCommitHooks = true
			}
		}
	}()

	err = do(ctx)
	return
}
