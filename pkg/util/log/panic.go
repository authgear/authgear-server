package log

import (
	"context"
	"errors"

	"github.com/authgear/authgear-server/pkg/util/errorutil"
)

func PanicValue(logger *Logger, err error) {
	if !errors.Is(err, context.Canceled) {
		logger.WithError(err).
			WithField("stack", errorutil.Callers(10000)).
			Error("panic occurred")
	}
}
