package log

import (
	"github.com/authgear/authgear-server/pkg/util/errorutil"
)

func PanicValue(logger *Logger, err error) {
	if !IgnoreError(err) {
		logger.WithError(err).
			WithField("stack", errorutil.Callers(10000)).
			Error("panic occurred")
	}
}
