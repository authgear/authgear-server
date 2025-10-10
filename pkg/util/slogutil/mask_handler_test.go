package slogutil

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"testing"
	"time"

	slogmulti "github.com/samber/slog-multi"
	. "github.com/smartystreets/goconvey/convey"
)

type unknownType struct {
	Value string
}

func TestAddMaskPatterns(t *testing.T) {
	Convey("AddMaskPatterns", t, func() {
		ctx := context.Background()

		Convey("should add patterns to empty context", func() {
			patterns := []MaskPattern{
				NewPlainMaskPattern("secret"),
				NewRegexMaskPattern(`\d+`),
			}
			newCtx := AddMaskPatterns(ctx, patterns)
			result := GetMaskPatterns(newCtx)
			So(len(result), ShouldEqual, 2)
		})

		Convey("should append patterns to existing context", func() {
			patterns1 := []MaskPattern{NewPlainMaskPattern("secret")}
			patterns2 := []MaskPattern{NewRegexMaskPattern(`\d+`)}

			ctx1 := AddMaskPatterns(ctx, patterns1)
			ctx2 := AddMaskPatterns(ctx1, patterns2)

			result := GetMaskPatterns(ctx2)
			So(len(result), ShouldEqual, 2)
		})

		Convey("should handle empty pattern slice", func() {
			patterns := []MaskPattern{}
			newCtx := AddMaskPatterns(ctx, patterns)
			result := GetMaskPatterns(newCtx)
			So(len(result), ShouldEqual, 0)
		})
	})
}

func TestGetMaskPatterns(t *testing.T) {
	Convey("GetMaskPatterns", t, func() {
		Convey("should return empty slice for empty context", func() {
			ctx := context.Background()
			result := GetMaskPatterns(ctx)
			So(len(result), ShouldEqual, 0)
		})

		Convey("should return patterns from context", func() {
			ctx := context.Background()
			patterns := []MaskPattern{NewPlainMaskPattern("secret")}
			ctx = AddMaskPatterns(ctx, patterns)
			result := GetMaskPatterns(ctx)
			So(len(result), ShouldEqual, 1)
		})
	})
}

func TestMaskHandlerOptions_maskString(t *testing.T) {
	Convey("MaskHandlerOptions.maskString", t, func() {
		options := MaskHandlerOptions{
			MaskPatterns: []MaskPattern{
				NewPlainMaskPattern("secret"),
				NewRegexMaskPattern(`\d+`),
			},
			Mask: "***",
		}

		Convey("should mask matching patterns", func() {
			result := options.maskString("secret123")
			So(result, ShouldEqual, "******")
		})

		Convey("should handle no matches", func() {
			result := options.maskString("no sensitive data")
			So(result, ShouldEqual, "no sensitive data")
		})

		Convey("should handle empty string", func() {
			result := options.maskString("")
			So(result, ShouldEqual, "")
		})
	})
}

func TestMaskHandlerOptions_maskAttr(t *testing.T) {
	Convey("MaskHandlerOptions.maskAttr", t, func() {
		options := MaskHandlerOptions{
			MaskPatterns: []MaskPattern{
				NewPlainMaskPattern("secret"),
			},
			Mask: "***",
		}
		ctx := context.Background()

		Convey("should mask string attributes", func() {
			attr := slog.String("key", "secret value")
			result := options.maskAttr(ctx, attr)
			So(result, ShouldResemble, slog.Attr{
				Key:   "key",
				Value: slog.StringValue("*** value"),
			})
		})

		Convey("should handle bool attributes", func() {
			attr := slog.Bool("key", true)
			result := options.maskAttr(ctx, attr)
			So(result, ShouldResemble, slog.Attr{
				Key:   "key",
				Value: slog.BoolValue(true),
			})
		})

		Convey("should handle int64 attributes", func() {
			attr := slog.Int64("key", 123)
			result := options.maskAttr(ctx, attr)
			So(result, ShouldResemble, slog.Attr{
				Key:   "key",
				Value: slog.Int64Value(123),
			})
		})

		Convey("should handle float64 attributes", func() {
			attr := slog.Float64("key", 3.14)
			result := options.maskAttr(ctx, attr)
			So(result, ShouldResemble, slog.Attr{
				Key:   "key",
				Value: slog.Float64Value(3.14),
			})
		})

		Convey("should handle uint64 attributes", func() {
			attr := slog.Uint64("key", 456)
			result := options.maskAttr(ctx, attr)
			So(result, ShouldResemble, slog.Attr{
				Key:   "key",
				Value: slog.Uint64Value(456),
			})
		})

		Convey("should handle time attributes", func() {
			now := time.Now()
			attr := slog.Time("key", now)
			result := options.maskAttr(ctx, attr)
			So(result, ShouldResemble, slog.Attr{
				Key:   "key",
				Value: slog.TimeValue(now),
			})
		})

		Convey("should handle duration attributes", func() {
			d := time.Second * 5
			attr := slog.Duration("key", d)
			result := options.maskAttr(ctx, attr)
			So(result, ShouldResemble, slog.Attr{
				Key:   "key",
				Value: slog.DurationValue(d),
			})
		})

		Convey("should mask error attributes", func() {
			err := errors.New("secret error")
			attr := slog.Any("key", err)
			result := options.maskAttr(ctx, attr)
			So(result.Key, ShouldEqual, "key")
			So(result.Value.Kind(), ShouldEqual, slog.KindAny)
			maskedValue, ok := result.Value.Any().(*MaskedError)
			So(ok, ShouldBeTrue)
			So(maskedValue.Type, ShouldEqual, "*errors.errorString")
			So(maskedValue.Message, ShouldEqual, "*** error")
		})

		Convey("should mask any attributes (unknown type) by converting to string", func() {
			attr := slog.Any("key", unknownType{Value: "secret value"})
			result := options.maskAttr(ctx, attr)
			So(result.Key, ShouldEqual, "key")
			So(result.Value.Kind(), ShouldEqual, slog.KindAny)
			maskedValue, ok := result.Value.Any().(*MaskedAny)
			So(ok, ShouldBeTrue)
			So(maskedValue.Type, ShouldEqual, "slogutil.unknownType")
			So(maskedValue.String(), ShouldEqual, "{*** value}")
		})

		Convey("should handle group attributes", func() {
			attr := slog.Group("group",
				slog.String("inner", "secret data"),
				slog.Int("count", 42),
			)
			result := options.maskAttr(ctx, attr)

			So(result, ShouldResemble, slog.Attr{
				Key: "group",
				Value: slog.GroupValue(
					slog.Attr{Key: "inner", Value: slog.StringValue("*** data")},
					slog.Attr{Key: "count", Value: slog.Int64Value(42)},
				),
			})
		})

		Convey("should handle LogValuer attributes", func() {
			valuer := slog.StringValue("secret message")
			attr := slog.Attr{Key: "key", Value: valuer}
			result := options.maskAttr(ctx, attr)
			So(result, ShouldResemble, slog.Attr{
				Key:   "key",
				Value: slog.StringValue("*** message"),
			})
		})
	})
}

func TestNewDefaultMaskHandlerOptions(t *testing.T) {
	Convey("NewDefaultMaskHandlerOptions", t, func() {
		options := NewDefaultMaskHandlerOptions()

		Convey("should have default mask patterns", func() {
			So(len(options.MaskPatterns), ShouldEqual, 2)
			So(options.Mask, ShouldEqual, "********")
		})

		Convey("should mask JWT tokens", func() {
			jwt := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c"
			result := options.maskString(jwt)
			So(result, ShouldEqual, "********")
		})

		Convey("should mask session tokens", func() {
			sessionToken := "12345678-1234-1234-1234-123456789012.AbCdEfGhIjKlMnOpQrStUvWxYz"
			result := options.maskString(sessionToken)
			So(result, ShouldEqual, "********")
		})
	})
}

func TestMaskHandler_Enabled(t *testing.T) {
	Convey("MaskHandler.Enabled", t, func() {
		handler := &MaskHandler{}

		Convey("should always return true", func() {
			So(handler.Enabled(context.Background(), slog.LevelDebug), ShouldBeTrue)
			So(handler.Enabled(context.Background(), slog.LevelInfo), ShouldBeTrue)
			So(handler.Enabled(context.Background(), slog.LevelWarn), ShouldBeTrue)
			So(handler.Enabled(context.Background(), slog.LevelError), ShouldBeTrue)
		})
	})
}

func TestMaskHandler_Handle(t *testing.T) {
	Convey("MaskHandler.Handle", t, func() {
		var w strings.Builder
		options := MaskHandlerOptions{
			MaskPatterns: []MaskPattern{
				NewPlainMaskPattern("secret"),
			},
			Mask: "***",
		}
		handler := &MaskHandler{
			Options: options,
			Next:    NewHandlerForTesting(slog.LevelInfo, &w),
		}

		Convey("should respect wrapped handler Enabled()", func() {
			ctx := context.Background()
			record := slog.NewRecord(time.Now(), slog.LevelDebug, "test message", 0)

			err := handler.Handle(ctx, record)
			So(err, ShouldBeNil)
			So(w.String(), ShouldEqual, "")
		})

		Convey("should mask attributes in log record", func() {
			ctx := context.Background()
			record := slog.NewRecord(time.Now(), slog.LevelInfo, "test message", 0)
			record.AddAttrs(slog.String("data", "secret value"))

			err := handler.Handle(ctx, record)
			So(err, ShouldBeNil)
			So(w.String(), ShouldContainSubstring, "data=\"*** value\"")
		})

		Convey("should combine patterns from options and context", func() {
			ctx := context.Background()
			ctx = AddMaskPatterns(ctx, []MaskPattern{
				NewPlainMaskPattern("password"),
			})

			record := slog.NewRecord(time.Now(), slog.LevelInfo, "test message", 0)
			record.AddAttrs(
				slog.String("secret_data", "secret info"),
				slog.String("auth_data", "password123"),
			)

			err := handler.Handle(ctx, record)
			So(err, ShouldBeNil)
			So(w.String(), ShouldContainSubstring, "secret_data=\"*** info\"")
			So(w.String(), ShouldContainSubstring, "auth_data=***123")
		})

		Convey("should handle empty context patterns", func() {
			ctx := context.Background()
			record := slog.NewRecord(time.Now(), slog.LevelInfo, "test message", 0)
			record.AddAttrs(slog.String("data", "secret value"))

			err := handler.Handle(ctx, record)
			So(err, ShouldBeNil)
			So(w.String(), ShouldContainSubstring, "data=\"*** value\"")
		})

		Convey("should handle multiple attributes", func() {
			ctx := context.Background()
			record := slog.NewRecord(time.Now(), slog.LevelInfo, "test message", 0)
			record.AddAttrs(
				slog.String("public", "safe data"),
				slog.String("private", "secret data"),
				slog.Int("count", 42),
			)

			err := handler.Handle(ctx, record)
			So(err, ShouldBeNil)
			So(w.String(), ShouldContainSubstring, "public=\"safe data\"")
			So(w.String(), ShouldContainSubstring, "private=\"*** data\"")
			So(w.String(), ShouldContainSubstring, "count=42")
		})
	})
}

func TestMaskHandler_WithAttrs(t *testing.T) {
	Convey("MaskHandler.WithAttrs", t, func() {
		var w strings.Builder
		options := MaskHandlerOptions{
			MaskPatterns: []MaskPattern{
				NewPlainMaskPattern("secret"),
			},
			Mask: "***",
		}
		handler := &MaskHandler{
			Options: options,
			Next:    NewHandlerForTesting(slog.LevelInfo, &w),
		}

		Convey("should return new handler with attributes", func() {
			attrs := []slog.Attr{slog.String("key", "value")}
			newHandler := handler.WithAttrs(attrs)

			// Should be a MaskHandler
			maskHandler, ok := newHandler.(*MaskHandler)
			So(ok, ShouldBeTrue)
			So(maskHandler.Options.Mask, ShouldEqual, "***")
		})
	})
}

func TestMaskHandler_WithGroup(t *testing.T) {
	Convey("MaskHandler.WithGroup", t, func() {
		var w strings.Builder
		options := MaskHandlerOptions{
			MaskPatterns: []MaskPattern{
				NewPlainMaskPattern("secret"),
			},
			Mask: "***",
		}
		handler := &MaskHandler{
			Options: options,
			Next:    NewHandlerForTesting(slog.LevelInfo, &w),
		}

		Convey("should return new handler with group", func() {
			newHandler := handler.WithGroup("group")

			// Should be a MaskHandler
			maskHandler, ok := newHandler.(*MaskHandler)
			So(ok, ShouldBeTrue)
			So(maskHandler.Options.Mask, ShouldEqual, "***")
		})
	})
}

func TestNewMaskMiddleware(t *testing.T) {
	Convey("NewMaskMiddleware", t, func() {
		var w strings.Builder
		options := MaskHandlerOptions{
			MaskPatterns: []MaskPattern{
				NewPlainMaskPattern("secret"),
			},
			Mask: "***",
		}

		logger := slog.New(slogmulti.Pipe(NewMaskMiddleware(options)).Handler(NewHandlerForTesting(slog.LevelInfo, &w)))

		Convey("should respect wrapped handler Enabled()", func() {
			logger.Debug("test message", slog.String("data", "secret value"))
			So(w.String(), ShouldContainSubstring, "")
		})

		Convey("should create working middleware", func() {
			logger.Info("test message", slog.String("data", "secret value"))
			So(w.String(), ShouldContainSubstring, "data=\"*** value\"")
		})

		Convey("should handle error logging", func() {
			err := fmt.Errorf("secret error occurred")
			logger.Error("error happened", slog.Any("error", err))
			So(w.String(), ShouldContainSubstring, "error=\"*** error occurred\"")
		})

		Convey("should work with context patterns", func() {
			ctx := context.Background()
			ctx = AddMaskPatterns(ctx, []MaskPattern{
				NewPlainMaskPattern("password"),
			})

			logger.InfoContext(ctx, "test message",
				slog.String("secret_data", "secret info"),
				slog.String("auth_data", "password123"))

			So(w.String(), ShouldContainSubstring, "secret_data=\"*** info\"")
			So(w.String(), ShouldContainSubstring, "auth_data=***123")
		})
	})
}

func TestMaskHandlerIntegration(t *testing.T) {
	Convey("MaskHandler Integration", t, func() {
		var w strings.Builder
		logger := slog.New(slogmulti.Pipe(NewMaskMiddleware(NewDefaultMaskHandlerOptions())).Handler(NewHandlerForTesting(slog.LevelInfo, &w)))

		Convey("should mask real-world sensitive data", func() {
			jwt := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c"
			sessionToken := "12345678-1234-1234-1234-123456789012.AbCdEfGhIjKlMnOpQrStUvWxYz"

			logger.Info("authentication",
				slog.String("jwt", jwt),
				slog.String("session", sessionToken),
				slog.String("user", "john.doe"))

			So(w.String(), ShouldContainSubstring, "jwt=********")
			So(w.String(), ShouldContainSubstring, "session=********")
			So(w.String(), ShouldContainSubstring, "user=john.doe")
		})

		Convey("should handle complex error with sensitive data", func() {
			err := fmt.Errorf("authentication failed with token: eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c")

			logger.Error("auth error", slog.Any("error", err))
			So(w.String(), ShouldContainSubstring, "error=\"authentication failed with token: ********\"")
		})

		Convey("should mask data attached with With()", func() {
			jwt := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c"
			sessionToken := "12345678-1234-1234-1234-123456789012.AbCdEfGhIjKlMnOpQrStUvWxYz"

			logger.With(
				slog.String("jwt", jwt),
				slog.String("session", sessionToken),
				slog.String("user", "john.doe"),
			).Info("authentication")

			So(w.String(), ShouldContainSubstring, "jwt=********")
			So(w.String(), ShouldContainSubstring, "session=********")
			So(w.String(), ShouldContainSubstring, "user=john.doe")
		})

		Convey("should mask data nested in group", func() {
			jwt := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c"
			sessionToken := "12345678-1234-1234-1234-123456789012.AbCdEfGhIjKlMnOpQrStUvWxYz"

			logger.WithGroup("g").With(
				slog.String("jwt", jwt),
				slog.String("session", sessionToken),
				slog.String("user", "john.doe"),
			).Info("authentication")

			So(w.String(), ShouldContainSubstring, "g.jwt=********")
			So(w.String(), ShouldContainSubstring, "g.session=********")
			So(w.String(), ShouldContainSubstring, "g.user=john.doe")
		})
	})
}
