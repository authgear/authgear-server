package logging

import (
	"net/http"

	"github.com/sirupsen/logrus"

	corehttp "github.com/skygeario/skygear-server/pkg/core/http"
)

func NewFactory(formatter logrus.Formatter) Factory {
	return Factory{formatter: formatter}
}

func NewFactoryFromRequest(r *http.Request, formatter logrus.Formatter) Factory {
	return Factory{
		requestID: r.Header.Get(corehttp.HeaderRequestID),
		formatter: formatter,
	}
}

func NewFactoryFromRequestID(requestID string, formatter logrus.Formatter) Factory {
	return Factory{
		requestID: requestID,
		formatter: formatter,
	}
}

type Factory struct {
	requestID string
	formatter logrus.Formatter
}

func (f Factory) NewLogger(name string) *logrus.Entry {
	entry := LoggerEntry(name)
	if f.requestID != "" {
		entry = entry.WithField("request_id", f.requestID)
	}
	entry.Logger.Formatter = f.formatter
	return entry
}
