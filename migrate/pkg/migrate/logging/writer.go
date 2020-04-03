package logging

import (
	"github.com/sirupsen/logrus"
)

type LogWriter struct {
	logger *logrus.Logger
	level  logrus.Level
}

func NewLogWriter(l *logrus.Logger, level logrus.Level) *LogWriter {
	lw := &LogWriter{}
	lw.logger = l
	lw.level = level
	return lw
}

func (w LogWriter) Write(p []byte) (n int, err error) {
	w.logger.Log(w.level, string(p))
	return len(p), nil
}
