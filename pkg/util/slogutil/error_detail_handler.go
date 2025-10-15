package slogutil

import (
	"context"
	"fmt"
	"log/slog"

	slogmulti "github.com/samber/slog-multi"

	"github.com/authgear/authgear-server/pkg/util/errorutil"
)

type ErrorDetailHandler struct {
	Next    slog.Handler
	Details errorutil.Details
}

var _ slog.Handler = (*StackTraceHandler)(nil)

func NewErrorDetailMiddleware() slogmulti.Middleware {
	return func(next slog.Handler) slog.Handler {
		return &ErrorDetailHandler{
			Next: next,
		}
	}
}

func (s *ErrorDetailHandler) Enabled(context.Context, slog.Level) bool {
	return true
}

func (s *ErrorDetailHandler) Handle(ctx context.Context, record slog.Record) error {
	details := s.Details

	recordAttrs := []slog.Attr{}
	record.Attrs(func(a slog.Attr) bool {
		recordAttrs = append(recordAttrs, a)
		return true
	})

	recordDetail := s.collectDetails(recordAttrs)
	if recordDetail != nil {
		details = recordDetail
	}

	for k, v := range details {
		record.AddAttrs(slog.Attr{
			Key:   fmt.Sprintf("details.%s", k),
			Value: slog.AnyValue(v),
		})
	}

	if s.Next.Enabled(ctx, record.Level) {
		return s.Next.Handle(ctx, record)
	}
	return nil
}

func (s *ErrorDetailHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	details := s.collectDetails(attrs)
	if details != nil {
		return &ErrorDetailHandler{
			Next:    s.Next.WithAttrs(attrs),
			Details: details,
		}
	}
	return &ErrorDetailHandler{
		Next:    s.Next.WithAttrs(attrs),
		Details: s.Details,
	}
}

func (s *ErrorDetailHandler) collectDetails(attrs []slog.Attr) errorutil.Details {
	var err error
	for _, attr := range attrs {
		if attr.Key == AttrKeyError {
			if maybeErr, ok := attr.Value.Any().(error); ok {
				err = maybeErr
				break
			}
		}
	}

	if err == nil {
		return nil
	}

	return errorutil.CollectDetails(err, nil)
}

func (s *ErrorDetailHandler) WithGroup(name string) slog.Handler {
	return &ErrorDetailHandler{
		Next: s.Next.WithGroup(name),
	}
}
