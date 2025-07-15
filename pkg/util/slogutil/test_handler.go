package slogutil

import (
	"io"
	"log/slog"
)

// NewHandlerForTesting is a wrapper around slog.NewTextHandler.
// It sets level to Debug, and removes the time key.
func NewHandlerForTesting(w io.Writer) slog.Handler {
	return slog.NewTextHandler(w, &slog.HandlerOptions{
		// For testing purpose, log everything.
		Level: slog.LevelDebug,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			// For testing purpose, omit the time key.
			if a.Key == slog.TimeKey {
				return slog.Attr{}
			}
			return a
		},
	})
}

func NewHandlerForTestingWithSource(w io.Writer) slog.Handler {
	return slog.NewTextHandler(w, &slog.HandlerOptions{
		// For testing purpose, log everything.
		Level: slog.LevelDebug,
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
