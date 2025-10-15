package slogutil

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"testing"

	slogmulti "github.com/samber/slog-multi"
	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/util/errorutil"
)

func TestNewErrorDetailMiddleware(t *testing.T) {
	Convey("NewErrorDetailMiddleware", t, func() {
		var w strings.Builder
		logger := slog.New(slogmulti.Pipe(
			NewErrorDetailMiddleware(),
		).Handler(NewHandlerForTesting(slog.LevelInfo, &w)))

		ctx := context.Background()

		Convey("log simple error", func() {
			w.Reset()
			err := fmt.Errorf("simple error")
			logger.With(AttrKeyError, err).InfoContext(ctx, "test")
			So(w.String(), ShouldEqual, "level=INFO msg=test error=\"simple error\"\n")
		})

		Convey("log error with details", func() {
			w.Reset()
			err := errorutil.WithDetails(fmt.Errorf("simple error"), errorutil.Details{"a": 1})
			logger.With(AttrKeyError, err).InfoContext(ctx, "test")
			So(w.String(), ShouldEqual, "level=INFO msg=test error=\"simple error\" details.a=1\n")
		})
	})
}
