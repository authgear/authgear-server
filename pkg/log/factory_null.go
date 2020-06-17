package log

import (
	"io/ioutil"

	"github.com/sirupsen/logrus"
)

type nullLoggerFactory struct{}

func NewNullFactory() Factory { return nullLoggerFactory{} }

func (f nullLoggerFactory) NewLogger(name string) *logrus.Entry {
	return logrus.NewEntry(&logrus.Logger{Out: ioutil.Discard, Formatter: f})
}

func (f nullLoggerFactory) Format(*logrus.Entry) ([]byte, error) {
	return nil, nil
}
