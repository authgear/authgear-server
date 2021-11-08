package log

import (
	"context"
	"database/sql"
	"errors"
	"syscall"

	"github.com/lib/pq"

	"github.com/authgear/authgear-server/pkg/util/errorutil"
)

func PanicValue(logger *Logger, err error) {
	skip := false

	// In most case, when the HTTP request was canceled,
	// the error would be context.Canceled.
	// However, by observation, other errors may occur too.
	// We consider all of them as request canceled and do not log them.

	if errors.Is(err, context.Canceled) {
		skip = true
	}

	var pqError *pq.Error
	if errors.As(err, &pqError) {
		// https://www.postgresql.org/docs/13/errcodes-appendix.html
		// 57014 is query_canceled
		if pqError.Code == "57014" {
			skip = true
		}
	}

	if errors.Is(err, syscall.EPIPE) {
		// syscall.EPIPE is "write: broken pipe"
		skip = true
	}

	if errors.Is(err, sql.ErrTxDone) {
		// The sql package will rollback the transaction when the context is canceled.
		// https://pkg.go.dev/database/sql/driver#ConnBeginTx
		// So when we call Rollback() again, this error will be returned.
		skip = true
	}

	if !skip {
		logger.WithError(err).
			WithField("stack", errorutil.Callers(10000)).
			Error("panic occurred")
	}
}
