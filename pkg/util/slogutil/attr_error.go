package slogutil

import (
	"log/slog"
)

// The stdlib does not include a way to include an error in the record.
// Hence we have this abstraction.

const AttrKeyError = "error"

func Err(err error) slog.Attr {
	return slog.Any(AttrKeyError, err)
}

func WithErr(logger *slog.Logger, err error) *slog.Logger {
	return logger.With(Err(err))
}
