package logging

import (
	"github.com/sirupsen/logrus"
)

type Factory interface {
	NewLogger(name string) *logrus.Entry
}

func NewFactory(hooks ...logrus.Hook) Factory {
	return factoryImpl{hooks: hooks}
}

type factoryImpl struct {
	hooks []logrus.Hook
}

func (f factoryImpl) NewLogger(name string) *logrus.Entry {
	entry := LoggerEntry(name)
	for _, hook := range f.hooks {
		entry.Logger.Hooks.Add(hook)
	}
	return entry
}
