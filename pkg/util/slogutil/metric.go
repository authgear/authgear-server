package slogutil

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"

	"github.com/lib/pq"
)

// MetricErrorName is a symbolic name for some errors
type MetricErrorName string

const (
	MetricErrorNameContextCanceled         MetricErrorName = "context.canceled"
	MetricErrorNameContextDeadlineExceeded MetricErrorName = "context.deadline_exceeded"
	MetricErrorNamePQ57014                 MetricErrorName = "pq.57014"
	MetricErrorNameSQLTxDone               MetricErrorName = "sql.tx_done"
	MetricErrorNameSQLDriverBadConn        MetricErrorName = "sql.driver.bad_conn"
)

func GetMetricErrorName(err error) (MetricErrorName, bool) {
	var pqError *pq.Error
	switch {
	case errors.Is(err, context.Canceled):
		// In most case, when the HTTP request was canceled,
		// the error would be context.Canceled.
		// However, by observation, other errors may occur too.
		// We consider all of them as request canceled and do not log them.
		return MetricErrorNameContextCanceled, true
	case errors.Is(err, context.DeadlineExceeded):
		// DeadlineExceeded occurs when we shutdown the daemon but encounter a timeout.
		return MetricErrorNameContextDeadlineExceeded, true
	case errors.As(err, &pqError):
		// https://www.postgresql.org/docs/13/errcodes-appendix.html
		// 57014 is query_canceled
		if pqError.Code == "57014" {
			return MetricErrorNamePQ57014, true
		}
	case errors.Is(err, sql.ErrTxDone):
		// The sql package will rollback the transaction when the context is canceled.
		// https://pkg.go.dev/database/sql/driver#ConnBeginTx
		// So when we call Rollback() again, this error will be returned.
		return MetricErrorNameSQLTxDone, true
	case errors.Is(err, driver.ErrBadConn):
		return MetricErrorNameSQLDriverBadConn, true
	}

	return "", false
}
