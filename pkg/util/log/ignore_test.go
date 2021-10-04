package log

import (
	"context"
	"fmt"
	"testing"

	"github.com/sirupsen/logrus"

	. "github.com/smartystreets/goconvey/convey"
)

func TestIgnore(t *testing.T) {
	Convey("Ignore", t, func() {

		Convey("Ignore entry with context.Canceled error", func() {
			logger := logrus.New()
			entry := logrus.NewEntry(logger)
			entry = entry.WithError(context.Canceled)

			So(Ignore(entry), ShouldBeTrue)
		})

		Convey("Ignore entry with wrapped context.Canceled error", func() {
			logger := logrus.New()
			entry := logrus.NewEntry(logger)

			entry = entry.WithError(fmt.Errorf("wrap: %w", context.Canceled))

			So(Ignore(entry), ShouldBeTrue)
		})

		Convey("Do not ignore any other entry", func() {
			logger := logrus.New()
			entry := logrus.NewEntry(logger)

			So(Ignore(entry), ShouldBeFalse)
		})
	})
}
