package slogutil

import (
	"log/slog"
	"os"
	"time"

	"github.com/lmittmann/tint"
	"github.com/mattn/go-isatty"
)

func NewStderrHandler(strLevel string) slog.Handler {
	level := ParseLevel(strLevel)
	addSource := level <= slog.LevelDebug
	w := os.Stderr

	isTerminal := isatty.IsTerminal(w.Fd())
	if isTerminal {
		return tint.NewHandler(w, &tint.Options{
			Level:      level,
			AddSource:  addSource,
			TimeFormat: time.RFC3339,
		})
	}

	return slog.NewTextHandler(w, &slog.HandlerOptions{
		Level:     level,
		AddSource: addSource,
	})
}
