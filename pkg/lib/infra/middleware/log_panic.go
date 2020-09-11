package middleware

import (
	"fmt"
	"net/http"

	"github.com/authgear/authgear-server/pkg/api"
	"github.com/authgear/authgear-server/pkg/util/errorutil"
	"github.com/authgear/authgear-server/pkg/util/log"
)

type LogPanicMiddlewareLogger struct{ *log.Logger }

func NewLogPanicMiddlewareLogger(lf *log.Factory) LogPanicMiddlewareLogger {
	return LogPanicMiddlewareLogger{lf.New("log-panic")}
}

type LogPanicMiddleware struct {
	Logger LogPanicMiddlewareLogger
}

func (m *LogPanicMiddleware) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				const errorTypeUnexpected = 0
				const errorTypeHandled = 1
				errorType := errorTypeUnexpected
				if herr, ok := err.(api.HandledError); ok {
					errorType = errorTypeHandled
					err = herr.Error
				}

				var e error
				if ee, isErr := err.(error); isErr {
					e = ee
				} else {
					e = fmt.Errorf("%+v", err)
				}

				if errorType == errorTypeUnexpected {
					m.Logger.WithError(e).
						WithField("stack", errorutil.Callers(10000)).
						Error("panic occurred")
				} else if errorType == errorTypeHandled {
					m.Logger.WithError(e).Error("unexpected error occurred")
				}

				// Rethrow
				panic(err)
			}
		}()

		next.ServeHTTP(w, r)

	})
}
