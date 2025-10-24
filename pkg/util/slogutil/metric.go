package slogutil

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"net"
	"os"

	"github.com/lib/pq"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"

	"github.com/authgear/authgear-server/pkg/util/otelutil"
)

// MetricOptionAttributeKeyValue is a MetricOption that adds an attribute.
type MetricOptionAttributeKeyValue struct {
	attribute.KeyValue
}

// ToOtelMetricOption implements otelutil.MetricOption.
func (o MetricOptionAttributeKeyValue) ToOtelMetricOption() metric.MeasurementOption {
	return metric.WithAttributes(o.KeyValue)
}

// MetricErrorName is a symbolic name for some errors
type MetricErrorName string

const (
	MetricErrorNameContextCanceled         MetricErrorName = "context.canceled"
	MetricErrorNameContextDeadlineExceeded MetricErrorName = "context.deadline_exceeded"
	MetricErrorNameOSErrDeadlineExceeded   MetricErrorName = "os.err_deadline_exceeded"
	MetricErrorNamePQ57014                 MetricErrorName = "pq.57014"
	MetricErrorNameSQLTxDone               MetricErrorName = "sql.tx_done"
	MetricErrorNameSQLDriverBadConn        MetricErrorName = "sql.driver.bad_conn"
	MetricErrorNameNetOpError              MetricErrorName = "net.op_error"
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
	case errors.Is(err, os.ErrDeadlineExceeded):
		return MetricErrorNameOSErrDeadlineExceeded, true
	case isNetOpError(err):
		// We used to identify syscall.ECONNRESET separately.
		// But I checked the log and found that syscall.ECONNRESET
		// was actually wrapped inside a *net.OpError.
		// Now that we track *net.OpError as metric,
		// there is no point in handling syscall.ECONNRESET specifically.
		return MetricErrorNameNetOpError, true
	}

	return "", false
}

func isNetOpError(err error) bool {
	var netOpError *net.OpError
	return errors.As(err, &netOpError)
}

func MetricOptionsForError(err error) []otelutil.MetricOption {
	var opts []otelutil.MetricOption
	if errorName, ok := GetMetricErrorName(err); ok {
		opts = append(opts, MetricOptionAttributeKeyValue{attribute.Key("error_name").String(string(errorName))})
	}

	var netOpError *net.OpError
	if errors.As(err, &netOpError) {
		opts = append(opts, MetricOptionAttributeKeyValue{attribute.Key("net.op").String(netOpError.Op)})
		opts = append(opts, MetricOptionAttributeKeyValue{attribute.Key("net.net").String(netOpError.Net)})
	}

	return opts
}
