package slogutil

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestAttrError(t *testing.T) {
	Convey("AttrError", t, func() {
		var w strings.Builder
		logger := slog.New(NewHandlerForTesting(slog.LevelError, &w))

		ctx := context.Background()

		Convey("log the attr error with its value", func() {
			err := fmt.Errorf("some error")
			WithErr(logger, err).ErrorContext(ctx, "testing")

			So(w.String(), ShouldEqual, "level=ERROR msg=testing error=\"some error\"\n")
		})

		Convey("log the attr error even it is nil", func() {
			WithErr(logger, nil).ErrorContext(ctx, "testing")

			So(w.String(), ShouldEqual, "level=ERROR msg=testing error=<nil>\n")
		})

		Convey("log the attr error when the error is wrapped", func() {
			err := fmt.Errorf("base error")
			err = fmt.Errorf("wrap error: %w", err)

			WithErr(logger, err).ErrorContext(ctx, "testing")
			So(w.String(), ShouldEqual, "level=ERROR msg=testing error=\"wrap error: base error\"\n")
		})

		Convey("log the attr error when the error is chained", func() {
			err := errors.Join(fmt.Errorf("error a"), fmt.Errorf("error b"))

			WithErr(logger, err).ErrorContext(ctx, "testing")
			So(w.String(), ShouldEqual, "level=ERROR msg=testing error=\"error a\\nerror b\"\n")
		})
	})
}
