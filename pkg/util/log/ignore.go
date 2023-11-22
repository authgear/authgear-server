package log

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"syscall"

	"github.com/lib/pq"
	"github.com/sirupsen/logrus"
)

func IgnoreEntry(entry *logrus.Entry) bool {
	if err, ok := entry.Data[logrus.ErrorKey].(error); ok {
		if IgnoreError(err) {
			return true
		}
	}

	return false
}

func IgnoreError(err error) (ignore bool) {
	// In most case, when the HTTP request was canceled,
	// the error would be context.Canceled.
	// However, by observation, other errors may occur too.
	// We consider all of them as request canceled and do not log them.

	if errors.Is(err, context.Canceled) {
		ignore = true
	}

	// ErrAbortHandler is a sentinel panic value to abort a handler.
	// net/http does NOT log this error, so do we.
	if errors.Is(err, http.ErrAbortHandler) {
		ignore = true
	}

	var pqError *pq.Error
	if errors.As(err, &pqError) {
		// https://www.postgresql.org/docs/13/errcodes-appendix.html
		// 57014 is query_canceled
		if pqError.Code == "57014" {
			ignore = true
		}
	}

	if errors.Is(err, syscall.EPIPE) {
		// syscall.EPIPE is "write: broken pipe"
		ignore = true
	}

	if errors.Is(err, sql.ErrTxDone) {
		// The sql package will rollback the transaction when the context is canceled.
		// https://pkg.go.dev/database/sql/driver#ConnBeginTx
		// So when we call Rollback() again, this error will be returned.
		ignore = true
	}

	return
}
