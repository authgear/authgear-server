package log

import (
	"io/ioutil"

	"github.com/sirupsen/logrus"
)

type nullFormatter struct{}

func (nullFormatter) Format(*logrus.Entry) ([]byte, error) {
	return nil, nil
}

var Null = logrus.NewEntry(&logrus.Logger{Out: ioutil.Discard, Formatter: nullFormatter{}})
