package apierrors

import (
	"fmt"
	"testing"

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
		logger.AddHook(log.SkipLoggingHook{})
		logger.AddHook(myhook)

		Convey("Ignore apierrors with some specific kind", func() {
			err := InternalError.WithReason("Ignore").SkipLoggingToExternalService().New("ignore")
			logger.WithError(err).Error("error")
			So(myhook.IsSkipped, ShouldBeTrue)
		})

		Convey("Ignore wrapped apierrors", func() {
			err := InternalError.WithReason("Ignore").SkipLoggingToExternalService().New("ignore")
			err = fmt.Errorf("wrap: %w", err)
			logger.WithError(err).Error("error")
			So(myhook.IsSkipped, ShouldBeTrue)
		})

		Convey("Do not apierrors", func() {
			err := InternalError.WithReason("DO_NOT_IGNORE").New("DO_NOT_IGNORE")
			logger.WithError(err).Error("error")
			So(myhook.IsSkipped, ShouldBeFalse)
		})
	})
}
