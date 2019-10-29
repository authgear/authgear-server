package logging

import (
	"net/http"

	"github.com/sirupsen/logrus"

	corehttp "github.com/skygeario/skygear-server/pkg/core/http"
)

type Factory interface {
	NewLogger(name string) *logrus.Entry
}

func NewFactory(hooks ...logrus.Hook) Factory {
	return factoryImpl{hooks: hooks}
}

func NewFactoryFromRequest(r *http.Request, hooks ...logrus.Hook) Factory {
	return factoryImpl{
		requestID: r.Header.Get(corehttp.HeaderRequestID),
		hooks:     hooks,
	}
}

func NewFactoryFromRequestID(requestID string, hooks ...logrus.Hook) Factory {
	return factoryImpl{
		requestID: requestID,
		hooks:     hooks,
	}
}

type factoryImpl struct {
	requestID string
	hooks     []logrus.Hook
}

func (f factoryImpl) NewLogger(name string) *logrus.Entry {
	entry := LoggerEntry(name)
	if f.requestID != "" {
		entry = entry.WithField("request_id", f.requestID)
	}
	for _, hook := range f.hooks {
		entry.Logger.Hooks.Add(hook)
	}
	return entry
}
