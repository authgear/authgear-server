package log

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/lib/pq"
	"github.com/sirupsen/logrus"

	"github.com/authgear/authgear-server/pkg/util/errorutil"
)

// This file defines a mechanism to skip logging.
// The previous mechanism requires Data[ErrorKey] to be of type error.
// But that mechanism conflicts with the masking hook.
// The masking hook changes Data[ErrorKey] to be a string, causing
// the previous mechanism to fail.

const KeySkipLogging = "__authgear_skip_logging"

func SkipLogging(e *logrus.Entry) {
	e.Data[KeySkipLogging] = true
}

func IsLoggingSkipped(e *logrus.Entry) bool {
	if b, ok := e.Data[KeySkipLogging].(bool); ok {
		return b
	}
	return false
}

type LoggingSkippable interface{ SkipLogging() bool }

type SkipLoggingHook struct{}

func (SkipLoggingHook) Levels() []logrus.Level { return logrus.AllLevels }

func (SkipLoggingHook) Fire(entry *logrus.Entry) error {
	if err, ok := entry.Data[logrus.ErrorKey].(error); ok {
		if IgnoreError(err) {
			SkipLogging(entry)
		}
	}

	return nil
}

func IgnoreError(err error) (ignore bool) {
	// ForceLogging overrides everything.
	if errorutil.IsForceLogging(err) {
		ignore = false
		return
	}

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

	// http.MaxBytesReader will *http.MaxBytesError when the body is too large.
	// We do not want to log this.
	var maxBytesError *http.MaxBytesError
	if errors.As(err, &maxBytesError) {
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

	if errors.Is(err, sql.ErrTxDone) {
		// The sql package will rollback the transaction when the context is canceled.
		// https://pkg.go.dev/database/sql/driver#ConnBeginTx
		// So when we call Rollback() again, this error will be returned.
		ignore = true
	}

	var skippable LoggingSkippable
	if errors.As(err, &skippable) {
		if skippable.SkipLogging() {
			ignore = true
		}
	}

	return
}
