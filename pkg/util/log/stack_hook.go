package log

import (
	"github.com/sirupsen/logrus"

	"github.com/authgear/authgear-server/pkg/util/errorutil"
)

var StackHookLevels = []logrus.Level{
	logrus.PanicLevel,
	logrus.FatalLevel,
	logrus.ErrorLevel,
}

// StackHook attaches call stack to entries with level >= error.
type StackHook struct{}

func (h *StackHook) Levels() []logrus.Level {
	return StackHookLevels
}

func (h *StackHook) Fire(entry *logrus.Entry) error {
	entry.Data["stack"] = errorutil.Callers(10000)
	return nil
}
