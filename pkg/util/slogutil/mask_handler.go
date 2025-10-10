package slogutil

import (
	"context"
	"log/slog"
	"reflect"
	"slices"

	"github.com/jba/slog/withsupport"
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

type MaskedValue interface {
	MaskedValue()
}

type MaskedError struct {
	Type    string
	Message string
}

func (e *MaskedError) MaskedValue() {}
func (e *MaskedError) Error() string {
	return e.Message
}

type MaskedAny struct {
	Type string
	str  string
}

func (e *MaskedAny) MaskedValue() {}
func (a *MaskedAny) String() string {
	return a.str
}

func (o MaskHandlerOptions) maskAttr(ctx context.Context, a slog.Attr) slog.Attr {
	switch a.Value.Kind() {
	case slog.KindAny:
		anything := a.Value.Any()
		if err, ok := anything.(error); ok {
			errSummary := errorutil.Summary(err)
			return slog.Attr{
				Key: a.Key,
				Value: slog.AnyValue(&MaskedError{
					Type:    reflect.TypeOf(err).String(),
					Message: o.maskString(errSummary)},
				),
			}
		}

		// Otherwise coerce to string and mask.
		str := o.maskString(a.Value.String())
		return slog.Attr{
			Key: a.Key,
			Value: slog.AnyValue(&MaskedAny{
				Type: reflect.TypeOf(a.Value.Any()).String(),
				str:  str,
			}),
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
	groupOrAttrs *withsupport.GroupOrAttrs
	Options      MaskHandlerOptions
	Next         slog.Handler
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

	// Gather a full list of attrs.
	attrs := []slog.Attr{}

	// Gather from captured With calls.
	attrs = append(attrs, LinearizeGroupOrAttrs(h.groupOrAttrs)...)
	// Gather from the record itself.
	record.Attrs(func(attr slog.Attr) bool {
		attrs = append(attrs, attr)
		return true
	})

	for idx, attr := range attrs {
		attrs[idx] = options.maskAttr(ctx, attr)
	}

	record = slog.NewRecord(record.Time, record.Level, record.Message, record.PC)
	record.AddAttrs(attrs...)

	if h.Next.Enabled(ctx, record.Level) {
		return h.Next.Handle(ctx, record)
	}
	return nil
}

func (h *MaskHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	// MaskHandler must capture the attrs in order to do masking.
	// So the call IS NOT forwarded to h.Next here.
	return &MaskHandler{
		groupOrAttrs: h.groupOrAttrs.WithAttrs(attrs),
		Options:      h.Options,
		Next:         h.Next,
	}
}

func (h *MaskHandler) WithGroup(name string) slog.Handler {
	// MaskHandler must capture the group in order to do masking.
	// So the call IS NOT forwarded to h.Next here.
	return &MaskHandler{
		groupOrAttrs: h.groupOrAttrs.WithGroup(name),
		Options:      h.Options,
		Next:         h.Next,
	}
}
