package log

import (
	"context"
	"errors"

	"github.com/sirupsen/logrus"
)

// Ignore reports whether the entry should be logged.
func Ignore(entry *logrus.Entry) bool {
	if err, ok := entry.Data[logrus.ErrorKey].(error); ok {
		if errors.Is(err, context.Canceled) {
			return true
		}
	}

	return false
}
