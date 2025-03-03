package log

import (
	"io/ioutil"
	"os"

	"github.com/sirupsen/logrus"
)

type Logger = logrus.Entry

type Factory struct {
	Level         Level
	DefaultFields logrus.Fields
	Hooks         []logrus.Hook
	Logger        *Logger
}

func NewFactory(level Level, hooks ...logrus.Hook) *Factory {
	logger := logrus.New()
	logger.Level = level.Logrus()
	logger.Out = ioutil.Discard
	logger.Hooks.Add(&StackHook{})
	logger.Hooks.Add(&ContextCauseHook{})
	logger.Hooks.Add(NewWriterHook(os.Stderr))

	for _, hook := range hooks {
		logger.Hooks.Add(hook)
	}

	return &Factory{
		Level:         level,
		DefaultFields: logrus.Fields{},
		Logger:        logger.WithFields(logrus.Fields{}),
		Hooks:         hooks,
	}
}

func (f *Factory) ReplaceHooks(hooks ...logrus.Hook) *Factory {
	factory := NewFactory(f.Level, hooks...)
	for k, v := range f.DefaultFields {
		factory.DefaultFields[k] = v
	}
	return factory
}

func (f *Factory) New(name string) *Logger {
	fields := logrus.Fields{"logger": name}
	for k, v := range f.DefaultFields {
		fields[k] = v
	}
	return f.Logger.WithFields(fields)
}
