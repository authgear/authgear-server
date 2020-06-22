package middlewares

import (
	"net/http"

	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/errors"
	"github.com/skygeario/skygear-server/pkg/core/handler"
	"github.com/skygeario/skygear-server/pkg/log"
)

// RecoverMiddleware recover from panic
type RecoverMiddleware struct {
	LoggerFactory *log.Factory
}

func (m *RecoverMiddleware) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				logger := m.LoggerFactory.New("recovery")

				const errorTypeUnexpected = 0
				const errorTypeHandled = 1
				const errorTypeConflict = 2
				errorType := errorTypeUnexpected
				if herr, ok := err.(handler.HandledError); ok {
					errorType = errorTypeHandled
					err = herr.Error
				} else if err, ok := err.(error); ok && errors.Is(err, db.ErrWriteConflict) {
					errorType = errorTypeConflict
				}

				var e error
				if ee, isErr := err.(error); isErr {
					e = ee
				} else {
					e = errors.Newf("%+v", err)
				}

				if errorType == errorTypeUnexpected {
					logger.WithError(e).WithField("stack", errors.Callers(8)).Error("panic occurred")
					w.WriteHeader(http.StatusInternalServerError)
				} else if errorType == errorTypeHandled {
					logger.WithError(e).Error("unexpected error occurred")
				} else if errorType == errorTypeConflict {
					w.WriteHeader(http.StatusConflict)
				}
			}
		}()

		next.ServeHTTP(w, r)
	})
}
