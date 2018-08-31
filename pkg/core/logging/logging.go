package logging

import (
	"github.com/sirupsen/logrus"
)


// CreateLogger create log entry for logging
func CreateLogger(module string) *logrus.Entry {
	logger := logrus.New()
	// For debug use
	// logger.SetLevel(logrus.DebugLevel)
	return logger.WithField("module", module)
}
