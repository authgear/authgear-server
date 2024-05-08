package apierrors

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"syscall"
	"testing"

	"github.com/lib/pq"
	"github.com/sirupsen/logrus"
	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/util/log"
)

type myhook struct {
	IsSkipped bool
}

func (*myhook) Levels() []logrus.Level { return logrus.AllLevels }

func (h *myhook) Fire(entry *logrus.Entry) error {
	if log.IsLoggingSkipped(entry) {
		h.IsSkipped = true
	}

	return nil
}

func TestSkipLogging(t *testing.T) {
	Convey("SkipLogging", t, func() {
		myhook := &myhook{}
		logger := logrus.New()
		logger.AddHook(SkipLoggingHook{})
		logger.AddHook(myhook)

		Convey("Ignore entry with context.Canceled error", func() {
			logger.WithError(context.Canceled).Error("error")
			So(myhook.IsSkipped, ShouldBeTrue)
		})

		Convey("Ignore entry with context.DeadlineExceeded error", func() {
			logger.WithError(context.DeadlineExceeded).Error("error")
			So(myhook.IsSkipped, ShouldBeTrue)
		})

		Convey("Ignore entry with http.ErrAbortHandler error", func() {
			logger.WithError(http.ErrAbortHandler).Error("error")
			So(myhook.IsSkipped, ShouldBeTrue)
		})

		Convey("Ignore json.SyntaxError", func() {
			var anything interface{}
			err := json.Unmarshal([]byte("{"), &anything)
			So(err, ShouldBeError, "unexpected end of JSON input")

			logger.WithError(err).Error("error")
			So(myhook.IsSkipped, ShouldBeTrue)
		})

		Convey("Ignore pg.Error.Code 57014", func() {
			err := &pq.Error{
				Code: "57014",
			}

			logger.WithError(err).Error("error")
			So(myhook.IsSkipped, ShouldBeTrue)
		})

		Convey("Ignore syscall.EPIPE", func() {
			err := syscall.EPIPE
			logger.WithError(err).Error("error")
			So(err, ShouldBeError, "broken pipe")
			So(myhook.IsSkipped, ShouldBeTrue)
		})

		Convey("Ignore syscall.ECONNREFUSED", func() {
			err := syscall.ECONNREFUSED
			logger.WithError(err).Error("error")
			So(err, ShouldBeError, "connection refused")
			So(myhook.IsSkipped, ShouldBeTrue)
		})

		Convey("Ignore sql.ErrTxDone error", func() {
			logger.WithError(sql.ErrTxDone).Error("error")
			So(myhook.IsSkipped, ShouldBeTrue)
		})

		Convey("Ignore entry with wrapped error", func() {
			logger.WithError(fmt.Errorf("wrap: %w", context.Canceled)).Error("error")
			So(myhook.IsSkipped, ShouldBeTrue)
		})

		Convey("Do not ignore any other entry", func() {
			logger.Error("error")
			So(myhook.IsSkipped, ShouldBeFalse)
		})
	})
}
