package slogutil

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"net"
	"os"
	"syscall"
	"testing"

	"github.com/lib/pq"
	. "github.com/smartystreets/goconvey/convey"
)

func TestGetMetricErrorName(t *testing.T) {
	Convey("GetMetricErrorName", t, func() {
		Convey("should return context.canceled for context.Canceled", func() {
			err := context.Canceled
			name, ok := GetMetricErrorName(err)
			So(ok, ShouldBeTrue)
			So(name, ShouldEqual, MetricErrorNameContextCanceled)
		})

		Convey("should return context.canceled for wrapped context.Canceled", func() {
			wrappedErr := fmt.Errorf("wrapping: %w", context.Canceled)
			name, ok := GetMetricErrorName(wrappedErr)
			So(ok, ShouldBeTrue)
			So(name, ShouldEqual, MetricErrorNameContextCanceled)
		})

		Convey("should return context.deadline_exceeded for context.DeadlineExceeded", func() {
			err := context.DeadlineExceeded
			name, ok := GetMetricErrorName(err)
			So(ok, ShouldBeTrue)
			So(name, ShouldEqual, MetricErrorNameContextDeadlineExceeded)
		})

		Convey("should return context.deadline_exceeded for wrapped context.DeadlineExceeded", func() {
			wrappedErr := fmt.Errorf("wrapping: %w", context.DeadlineExceeded)
			name, ok := GetMetricErrorName(wrappedErr)
			So(ok, ShouldBeTrue)
			So(name, ShouldEqual, MetricErrorNameContextDeadlineExceeded)
		})

		Convey("should return pq.57014 for PostgreSQL query_canceled error", func() {
			pqErr := &pq.Error{Code: "57014"}
			name, ok := GetMetricErrorName(pqErr)
			So(ok, ShouldBeTrue)
			So(name, ShouldEqual, MetricErrorNamePQ57014)
		})

		Convey("should not match pq.Error with different code", func() {
			pqErr := &pq.Error{Code: "23505"}
			name, ok := GetMetricErrorName(pqErr)
			So(ok, ShouldBeFalse)
			So(name, ShouldEqual, MetricErrorName(""))
		})

		Convey("should return os.err_deadline_exceeded for os.ErrDeadlineExceeded", func() {
			err := os.ErrDeadlineExceeded
			name, ok := GetMetricErrorName(err)

			So(err, ShouldBeError, "i/o timeout") // This serves as a documentation of what os.ErrDeadlineExceeded is

			So(ok, ShouldBeTrue)
			So(name, ShouldEqual, MetricErrorNameOSErrDeadlineExceeded)
		})

		Convey("should return false for syscall.ECONNRESET", func() {
			err := syscall.ECONNRESET
			name, ok := GetMetricErrorName(err)
			So(ok, ShouldBeFalse)
			So(name, ShouldEqual, MetricErrorName(""))
		})

		Convey("should return false for other syscall errors", func() {
			err := syscall.ECONNREFUSED
			name, ok := GetMetricErrorName(err)
			So(ok, ShouldBeFalse)
			So(name, ShouldEqual, MetricErrorName(""))
		})

		Convey("should return sql.tx_done for sql.ErrTxDone", func() {
			err := sql.ErrTxDone
			name, ok := GetMetricErrorName(err)
			So(ok, ShouldBeTrue)
			So(name, ShouldEqual, MetricErrorNameSQLTxDone)
		})

		Convey("should return sql.driver.bad_conn for driver.ErrBadConn", func() {
			err := driver.ErrBadConn
			name, ok := GetMetricErrorName(err)
			So(ok, ShouldBeTrue)
			So(name, ShouldEqual, MetricErrorNameSQLDriverBadConn)
		})

		Convey("should return net.op_error for *net.OpError", func() {
			err := &net.OpError{
				Op:  "write",
				Net: "tcp",
				Err: syscall.ECONNRESET,
			}
			name, ok := GetMetricErrorName(err)
			So(ok, ShouldBeTrue)
			So(name, ShouldEqual, MetricErrorNameNetOpError)
		})

		Convey("should return false for unrecognized errors", func() {
			err := errors.New("some other error")
			name, ok := GetMetricErrorName(err)
			So(ok, ShouldBeFalse)
			So(name, ShouldEqual, MetricErrorName(""))
		})

		Convey("should return false for nil error", func() {
			name, ok := GetMetricErrorName(nil)
			So(ok, ShouldBeFalse)
			So(name, ShouldEqual, MetricErrorName(""))
		})
	})
}
