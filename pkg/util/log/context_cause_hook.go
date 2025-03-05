package log

import (
	"context"

	"github.com/sirupsen/logrus"
)

// ContextCauseHook attaches context.Cause.
type ContextCauseHook struct{}

func (h *ContextCauseHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

func (h *ContextCauseHook) Fire(entry *logrus.Entry) error {
	ctx := entry.Context

	// No context associated with this entry.
	if ctx == nil {
		entry.Data["context_cause"] = "<context-is-nil>"
	} else {
		ctxErr := ctx.Err()
		// ctx is not canceled.
		if ctxErr == nil {
			entry.Data["context_cause"] = "<context-err-is-nil>"
		} else {
			cause := context.Cause(ctx)
			// ctx is canceled but there is no cause.
			if cause == nil {
				entry.Data["context_cause"] = "<context-cause-is-nil>"
			} else {
				entry.Data["context_cause"] = cause.Error()
			}
		}
	}

	return nil
}
