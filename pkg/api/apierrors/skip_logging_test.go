package apierrors

import (
	"context"
	"encoding/json"
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
		logger.AddHook(SkipLoggingHook{})
		logger.AddHook(myhook)

		Convey("Ignore entry with context.Canceled error", func() {
			logger.WithError(context.Canceled).Error("error")
			So(myhook.IsSkipped, ShouldBeTrue)
		})

		Convey("Ignore entry with wrapped context.Canceled error", func() {
			logger.WithError(fmt.Errorf("wrap: %w", context.Canceled)).Error("error")
			So(myhook.IsSkipped, ShouldBeTrue)
		})

		Convey("Ignore json.SyntaxError", func() {
			var anything interface{}
			err := json.Unmarshal([]byte("{"), &anything)
			So(err, ShouldBeError, "unexpected end of JSON input")

			logger.WithError(err).Error("error")
			So(myhook.IsSkipped, ShouldBeTrue)
		})

		Convey("Do not ignore any other entry", func() {
			logger.Error("error")
			So(myhook.IsSkipped, ShouldBeFalse)
		})
	})
}
