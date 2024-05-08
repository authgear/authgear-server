package apierrors

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"syscall"

	"github.com/lib/pq"
	"github.com/sirupsen/logrus"

	"github.com/authgear/authgear-server/pkg/util/log"
)

var SkipLoggingForKinds map[Kind]bool

func init() {
	SkipLoggingForKinds = make(map[Kind]bool)
}

type SkipLoggingHook struct{}

func (SkipLoggingHook) Levels() []logrus.Level { return logrus.AllLevels }

func (SkipLoggingHook) Fire(entry *logrus.Entry) error {
	if err, ok := entry.Data[logrus.ErrorKey].(error); ok {
		if IgnoreError(err) {
			log.SkipLogging(entry)
		}
	}

	return nil
}

func IgnoreError(err error) (ignore bool) {
	// In most case, when the HTTP request was canceled,
	// the error would be context.Canceled.
	// However, by observation, other errors may occur too.
	// We consider all of them as request canceled and do not log them.
	if errors.Is(err, context.Canceled) {
		ignore = true
	}

	// DeadlineExceeded occurs when we shutdown the daemon but encounter a timeout.
	if errors.Is(err, context.DeadlineExceeded) {
		ignore = true
	}

	// ErrAbortHandler is a sentinel panic value to abort a handler.
	// net/http does NOT log this error, so do we.
	if errors.Is(err, http.ErrAbortHandler) {
		ignore = true
	}

	// json.Unmarshal returns a SyntaxError if the JSON can't be parsed.
	// https://pkg.go.dev/encoding/json#SyntaxError
	var jsonSyntaxError *json.SyntaxError
	if errors.As(err, &jsonSyntaxError) {
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
		// syscall.EPIPE is "broken pipe"
		ignore = true
	}

	if errors.Is(err, sql.ErrTxDone) {
		// The sql package will rollback the transaction when the context is canceled.
		// https://pkg.go.dev/database/sql/driver#ConnBeginTx
		// So when we call Rollback() again, this error will be returned.
		ignore = true
	}

	// Skip logging for some kinds.
	if apiError := AsAPIError(err); apiError != nil {
		skipped := SkipLoggingForKinds[apiError.Kind]
		if skipped {
			ignore = true
		}
	}

	return
}
