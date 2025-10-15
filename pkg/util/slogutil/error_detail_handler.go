package slogutil

import (
	"context"
	"log/slog"

	"github.com/jba/slog/withsupport"
	slogmulti "github.com/samber/slog-multi"

	"github.com/authgear/authgear-server/pkg/util/errorutil"
)

type ErrorDetailHandler struct {
	Next slog.Handler
	// See https://github.com/golang/example/blob/master/slog-handler-guide/README.md#the-withgroup-method
	groupOrAttrs *withsupport.GroupOrAttrs
}

var _ slog.Handler = (*ErrorDetailHandler)(nil)

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

	allAttrs := []slog.Attr{}
	record.Attrs(func(a slog.Attr) bool {
		allAttrs = append(allAttrs, a)
		return true
	})
	s.groupOrAttrs.Apply(func(groups []string, a slog.Attr) {
		allAttrs = append(allAttrs, a)
	})

	details := s.collectDetails(allAttrs)

	detailAttrs := []any{}
	for k, v := range details {
		detailAttrs = append(detailAttrs, slog.Attr{
			Key:   k,
			Value: slog.AnyValue(v),
		})
	}
	newRecord := record.Clone()
	newRecord.AddAttrs(slog.Group("details", detailAttrs...))

	if s.Next.Enabled(ctx, newRecord.Level) {
		return s.Next.Handle(ctx, newRecord)
	}
	return nil
}

func (s *ErrorDetailHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &ErrorDetailHandler{
		Next:         s.Next.WithAttrs(attrs),
		groupOrAttrs: s.groupOrAttrs.WithAttrs(attrs),
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
		Next:         s.Next.WithGroup(name),
		groupOrAttrs: s.groupOrAttrs.WithGroup(name),
	}
}
