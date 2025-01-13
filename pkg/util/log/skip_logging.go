package log

import (
	"github.com/sirupsen/logrus"
)

// This file defines a mechanism to skip logging.
// The previous mechanism requires Data[ErrorKey] to be of type error.
// But that mechanism conflicts with the masking hook.
// The masking hook changes Data[ErrorKey] to be a string, causing
// the previous mechanism to fail.

const KeySkipLogging = "__authgear_skip_logging"

func SkipLogging(e *logrus.Entry) {
	e.Data[KeySkipLogging] = true
}

func IsLoggingSkipped(e *logrus.Entry) bool {
	if b, ok := e.Data[KeySkipLogging].(bool); ok {
		return b
	}
	return false
}

type LoggingSkippable interface{ SkipLogging() bool }
