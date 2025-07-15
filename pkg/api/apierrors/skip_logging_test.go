package apierrors

import (
	"context"
	"fmt"
	"log/slog"
	"testing"

	slogmulti "github.com/samber/slog-multi"
	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/util/slogutil"
)

type myhandler struct {
	IsSkipped bool
}

var _ slog.Handler = (*myhandler)(nil)

// Enabled implements slog.Handler.
func (m *myhandler) Enabled(context.Context, slog.Level) bool {
	return true
}

// Handle implements slog.Handler.
func (m *myhandler) Handle(ctx context.Context, record slog.Record) error {
	if slogutil.IsLoggingSkipped(record) {
		m.IsSkipped = true
	}
	return nil
}

// WithAttrs implements slog.Handler.
func (m *myhandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return m
}

// WithGroup implements slog.Handler.
func (m *myhandler) WithGroup(name string) slog.Handler {
	return m
}

func TestSkipLogging(t *testing.T) {
	Convey("SkipLogging", t, func() {
		ctx := context.Background()
		handler := &myhandler{}
		rootLogger := slog.New(slogmulti.Pipe(slogutil.NewSkipLoggingMiddleware()).Handler(handler))
		ctx = slogutil.SetContextLogger(ctx, rootLogger)

		logger := slogutil.NewLogger("logger")

		Convey("Ignore apierrors with some specific kind", func() {
			logger := logger.GetLogger(ctx)
			err := InternalError.WithReason("Ignore").SkipLoggingToExternalService().New("ignore")
			logger.WithError(err).Error(ctx, "error")
			So(handler.IsSkipped, ShouldBeTrue)
		})

		Convey("Ignore wrapped apierrors", func() {
			logger := logger.GetLogger(ctx)
			err := InternalError.WithReason("Ignore").SkipLoggingToExternalService().New("ignore")
			err = fmt.Errorf("wrap: %w", err)
			logger.WithError(err).Error(ctx, "error")
			So(handler.IsSkipped, ShouldBeTrue)
		})

		Convey("Do not apierrors", func() {
			logger := logger.GetLogger(ctx)
			err := InternalError.WithReason("DO_NOT_IGNORE").New("DO_NOT_IGNORE")
			logger.WithError(err).Error(ctx, "error")
			So(handler.IsSkipped, ShouldBeFalse)
		})
	})
}
