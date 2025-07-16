package slogutil

import (
	"context"
	"log/slog"
	"slices"

	slogmulti "github.com/samber/slog-multi"

	"github.com/authgear/authgear-server/pkg/util/errorutil"
)

type contextKeyTypeMaskPatterns struct{}

var contextKeyMaskPatterns = contextKeyTypeMaskPatterns{}

func AddMaskPatterns(ctx context.Context, newPatterns []MaskPattern) context.Context {
	patterns := []MaskPattern{}
	if existingPatterns, ok := ctx.Value(contextKeyMaskPatterns).([]MaskPattern); ok {
		patterns = append(patterns, existingPatterns...)
	}
	patterns = append(patterns, newPatterns...)
	ctx = context.WithValue(ctx, contextKeyMaskPatterns, patterns)
	return ctx
}

func GetMaskPatterns(ctx context.Context) []MaskPattern {
	patterns, ok := ctx.Value(contextKeyMaskPatterns).([]MaskPattern)
	if !ok {
		return nil
	}
	return patterns
}

type MaskHandlerOptions struct {
	MaskPatterns []MaskPattern
	Mask         string
}

func (o MaskHandlerOptions) maskString(s string) string {
	for _, p := range o.MaskPatterns {
		s = p.Mask(s, o.Mask)
	}
	return s
}

func (o MaskHandlerOptions) maskAttr(ctx context.Context, a slog.Attr) slog.Attr {
	switch a.Value.Kind() {
	case slog.KindAny:
		anything := a.Value.Any()
		if err, ok := anything.(error); ok {
			errSummary := errorutil.Summary(err)
			return slog.Attr{
				Key:   a.Key,
				Value: slog.StringValue(o.maskString(errSummary)),
			}
		}

		// Otherwise coerce to string and mask.
		str := o.maskString(a.Value.String())
		return slog.Attr{
			Key:   a.Key,
			Value: slog.StringValue(str),
		}
	case slog.KindBool:
		return a
	case slog.KindDuration:
		return a
	case slog.KindFloat64:
		return a
	case slog.KindInt64:
		return a
	case slog.KindString:
		str := o.maskString(a.Value.String())
		return slog.Attr{
			Key:   a.Key,
			Value: slog.StringValue(str),
		}
	case slog.KindTime:
		return a
	case slog.KindUint64:
		return a
	case slog.KindGroup:
		groupAttrs := []slog.Attr{}
		for _, groupAttr := range a.Value.Group() {
			groupAttr = o.maskAttr(ctx, groupAttr)
			groupAttrs = append(groupAttrs, groupAttr)
		}
		return slog.Attr{
			Key:   a.Key,
			Value: slog.GroupValue(groupAttrs...),
		}
	case slog.KindLogValuer:
		value := a.Value.Resolve()
		return o.maskAttr(ctx, slog.Attr{
			Key:   a.Key,
			Value: value,
		})
	default:
		// By default, coerce to string and mask.
		str := o.maskString(a.Value.String())
		return slog.Attr{
			Key:   a.Key,
			Value: slog.StringValue(str),
		}
	}
}

func NewDefaultMaskHandlerOptions() MaskHandlerOptions {
	return MaskHandlerOptions{
		MaskPatterns: []MaskPattern{
			NewRegexMaskPattern(`eyJ[A-Za-z0-9-_=]+\.[A-Za-z0-9-_=]+\.?[A-Za-z0-9-_.+/=]*`),
			NewRegexMaskPattern(`[A-Fa-f0-9-]{36}\.[A-Za-z0-9]+`),
		},
		Mask: "********",
	}
}

type MaskHandler struct {
	Options MaskHandlerOptions
	Next    slog.Handler
}

var _ slog.Handler = (*MaskHandler)(nil)

func NewMaskMiddleware(options MaskHandlerOptions) slogmulti.Middleware {
	return func(next slog.Handler) slog.Handler {
		return &MaskHandler{
			Options: options,
			Next:    next,
		}
	}
}

func (h *MaskHandler) Enabled(context.Context, slog.Level) bool {
	// Masking is always enabled.
	return true
}

func (h *MaskHandler) Handle(ctx context.Context, record slog.Record) error {
	patterns := slices.Clone(h.Options.MaskPatterns)
	patternsFromContext := GetMaskPatterns(ctx)
	patterns = append(patterns, patternsFromContext...)

	options := MaskHandlerOptions{
		MaskPatterns: patterns,
		Mask:         h.Options.Mask,
	}

	attrs := []slog.Attr{}
	record.Attrs(func(a slog.Attr) bool {
		attrs = append(attrs, options.maskAttr(ctx, a))
		return true
	})
	record = slog.NewRecord(record.Time, record.Level, record.Message, record.PC)
	record.AddAttrs(attrs...)

	if h.Next.Enabled(ctx, record.Level) {
		return h.Next.Handle(ctx, record)
	}
	return nil
}

func (h *MaskHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &MaskHandler{
		Options: h.Options,
		Next:    h.Next.WithAttrs(attrs),
	}
}

func (h *MaskHandler) WithGroup(name string) slog.Handler {
	return &MaskHandler{
		Options: h.Options,
		Next:    h.Next.WithGroup(name),
	}
}
