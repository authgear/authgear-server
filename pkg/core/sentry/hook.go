package sentry

import (
	"context"
	"fmt"

	"github.com/getsentry/sentry-go"

	"github.com/sirupsen/logrus"
)

var LogHookLevels = []logrus.Level{
	logrus.PanicLevel,
	logrus.FatalLevel,
	logrus.ErrorLevel,
	logrus.WarnLevel,
}

type LogHook struct {
	Hub *sentry.Hub
}

func NewLogHookFromContext(ctx context.Context) *LogHook {
	return &LogHook{Hub: sentry.GetHubFromContext(ctx)}
}

func (h *LogHook) Levels() []logrus.Level { return LogHookLevels }

func (h *LogHook) Fire(entry *logrus.Entry) error {
	if h.Hub == nil {
		return nil
	}

	event := makeEvent(entry)
	h.Hub.CaptureEvent(event)
	return nil
}

func makeEvent(entry *logrus.Entry) *sentry.Event {
	event := sentry.NewEvent()
	switch entry.Level {
	case logrus.PanicLevel, logrus.FatalLevel:
		event.Level = sentry.LevelFatal
	case logrus.ErrorLevel:
		event.Level = sentry.LevelError
	case logrus.WarnLevel:
		event.Level = sentry.LevelWarning
	case logrus.InfoLevel:
		event.Level = sentry.LevelInfo
	case logrus.DebugLevel:
		event.Level = sentry.LevelDebug
	}
	event.Timestamp = entry.Time.Unix()
	event.Message = entry.Message

	var err string
	needStack := false
	data := map[string]interface{}{}
	for k, v := range entry.Data {
		switch k {
		case "error":
			err = fmt.Sprint(v)
		case "stack":
			needStack = true
		case "logger":
			event.Logger = fmt.Sprint(v)
		case "module":
			event.Tags["module"] = fmt.Sprint(v)
		default:
			data[k] = v
		}
	}

	exception := sentry.Exception{Value: err}
	if needStack {
		exception.Stacktrace = sentry.NewStacktrace()
	}
	event.Exception = []sentry.Exception{exception}
	event.Extra = data

	return event
}
