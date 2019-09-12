package logging

import (
	"net/http"

	"github.com/sirupsen/logrus"
)

func NewFactory(r *http.Request, formatter logrus.Formatter) Factory {
	return Factory{r: r, formatter: formatter}
}

type Factory struct {
	r         *http.Request
	formatter logrus.Formatter
}

func (f Factory) NewLogger(name string) *logrus.Entry {
	return CreateLogger(f.r, name, f.formatter)
}
