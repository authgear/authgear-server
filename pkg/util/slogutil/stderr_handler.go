package slogutil

import (
	"log/slog"
	"os"
)

func NewStderrHandler(strLevel string) *slog.TextHandler {
	level := ParseLevel(strLevel)
	return slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: level,
		// Include source when the global log level is debug or lower.
		AddSource: level <= slog.LevelDebug,
	})
}
