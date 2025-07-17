package slogutil

import (
	"io"
	"log/slog"
)

// NewHandlerForTesting is a wrapper around slog.NewTextHandler.
// It removes the time key.
func NewHandlerForTesting(level slog.Leveler, w io.Writer) slog.Handler {
	return slog.NewTextHandler(w, &slog.HandlerOptions{
		Level: level,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			// For testing purpose, omit the time key.
			if a.Key == slog.TimeKey {
				return slog.Attr{}
			}
			return a
		},
	})
}

func NewHandlerForTestingWithSource(level slog.Leveler, w io.Writer) slog.Handler {
	return slog.NewTextHandler(w, &slog.HandlerOptions{
		Level: level,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			// For testing purpose, omit the time key.
			if a.Key == slog.TimeKey {
				return slog.Attr{}
			}
			return a
		},
		AddSource: true,
	})
}
