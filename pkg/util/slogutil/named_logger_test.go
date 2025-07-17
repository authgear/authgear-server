package slogutil

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestNamedLogger(t *testing.T) {
	Convey("NamedLogger", t, func() {
		ctx := context.Background()
		var w strings.Builder
		rootLogger := slog.New(NewHandlerForTesting(slog.LevelDebug, &w))
		ctx = SetContextLogger(ctx, rootLogger)

		serviceALogger := NewLogger("service.a")
		serviceBLogger := NewLogger("service.b")

		Convey("simple logging", func() {
			logger := serviceALogger.GetLogger(ctx)

			logger.Debug(ctx, "debug")
			logger.Info(ctx, "info")
			logger.Warn(ctx, "warn")
			logger.Error(ctx, "error")

			So(w.String(), ShouldEqual, `level=DEBUG msg=debug logger=service.a
level=INFO msg=info logger=service.a
level=WARN msg=warn logger=service.a
level=ERROR msg=error logger=service.a
`)
		})

		Convey("simple logging with attrs", func() {
			logger := serviceALogger.GetLogger(ctx)

			logger.Debug(ctx, "debug", slog.String("attr", "debug"))
			logger.Info(ctx, "info", slog.String("attr", "info"))

			So(w.String(), ShouldEqual, `level=DEBUG msg=debug logger=service.a attr=debug
level=INFO msg=info logger=service.a attr=info
`)
		})

		Convey("logging errors with WithError()", func() {
			logger := serviceALogger.GetLogger(ctx)

			err := fmt.Errorf("something went wrong")
			logger.WithError(err).Error(ctx, "error")
			So(w.String(), ShouldEqual, "level=ERROR msg=error logger=service.a error=\"something went wrong\"\n")
		})

		Convey("derive logger with With()", func() {
			logger := serviceALogger.GetLogger(ctx)

			logger_ := logger.With(slog.Bool("derived", true))

			logger.Info(ctx, "logger")
			logger_.Info(ctx, "logger_")

			So(w.String(), ShouldEqual, `level=INFO msg=logger logger=service.a
level=INFO msg=logger_ logger=service.a derived=true
`)
		})

		Convey("inject attrs with SetContextLogger()", func() {
			httpHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				ctx := r.Context()
				logger := serviceALogger.GetLogger(ctx)
				logger.Info(ctx, "success")
			})

			middleware := func(next http.Handler) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					ctx := r.Context()
					logger := GetContextLogger(ctx)
					logger = logger.With("app", "myapp")
					ctx = SetContextLogger(ctx, logger)
					r = r.WithContext(ctx)
					next.ServeHTTP(w, r)
				})
			}

			h := middleware(httpHandler)

			// The request context MUST BE a context with logger.
			req := httptest.NewRequestWithContext(ctx, "GET", "/", nil)
			recorder := httptest.NewRecorder()
			h.ServeHTTP(recorder, req)

			So(w.String(), ShouldEqual, "level=INFO msg=success app=myapp logger=service.a\n")
		})

		Convey("different packages using different named loggers", func() {
			serviceA := func(ctx context.Context) {
				logger := serviceALogger.GetLogger(ctx)
				logger.Info(ctx, "service a")
			}

			serviceB := func(ctx context.Context) {
				logger := serviceBLogger.GetLogger(ctx)
				logger.Info(ctx, "service b")
			}

			httpHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				ctx := r.Context()
				serviceA(ctx)
				serviceB(ctx)
			})

			middleware := func(next http.Handler) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					ctx := r.Context()
					logger := GetContextLogger(ctx)
					logger = logger.With("app", "myapp")
					ctx = SetContextLogger(ctx, logger)
					r = r.WithContext(ctx)
					next.ServeHTTP(w, r)
				})
			}

			h := middleware(httpHandler)

			// The request context MUST BE a context with logger.
			req := httptest.NewRequestWithContext(ctx, "GET", "/", nil)
			recorder := httptest.NewRecorder()
			h.ServeHTTP(recorder, req)

			So(w.String(), ShouldEqual, `level=INFO msg="service a" app=myapp logger=service.a
level=INFO msg="service b" app=myapp logger=service.b
`)
		})
	})

	Convey("NamedLogger handles PC correctly", t, func() {
		ctx := context.Background()
		var w strings.Builder
		rootLogger := slog.New(NewHandlerForTestingWithSource(slog.LevelInfo, &w))
		ctx = SetContextLogger(ctx, rootLogger)

		logger := NewLogger("logger")

		Convey("source is correct", func() {
			logger := logger.GetLogger(ctx)

			logger.Info(ctx, "testing")

			re := regexp.MustCompile("source=.*pkg/util/slogutil/named_logger_test.go:")
			matches := re.FindAllString(w.String(), -1)
			So(len(matches), ShouldEqual, 1)
		})
	})
}
