package log

import (
	"context"
	"fmt"
	"testing"

	"github.com/sirupsen/logrus"

	. "github.com/smartystreets/goconvey/convey"
)

func TestIgnoreEntry(t *testing.T) {
	Convey("IgnoreEntry", t, func() {

		Convey("Ignore entry with context.Canceled error", func() {
			logger := logrus.New()
			entry := logrus.NewEntry(logger)
			entry = entry.WithError(context.Canceled)

			So(IgnoreEntry(entry), ShouldBeTrue)
		})

		Convey("Ignore entry with wrapped context.Canceled error", func() {
			logger := logrus.New()
			entry := logrus.NewEntry(logger)

			entry = entry.WithError(fmt.Errorf("wrap: %w", context.Canceled))

			So(IgnoreEntry(entry), ShouldBeTrue)
		})

		Convey("Do not ignore any other entry", func() {
			logger := logrus.New()
			entry := logrus.NewEntry(logger)

			So(IgnoreEntry(entry), ShouldBeFalse)
		})
	})
}
