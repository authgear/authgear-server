package middleware

import (
	"net/http"

	"github.com/sirupsen/logrus"

	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/errors"
	"github.com/skygeario/skygear-server/pkg/core/handler"
	"github.com/skygeario/skygear-server/pkg/core/logging"
	"github.com/skygeario/skygear-server/pkg/core/sentry"
)

// RecoverMiddleware recover from panic
type RecoverMiddleware struct {
}

func (m RecoverMiddleware) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// prepare context for storing tenant config
		r = r.WithContext(config.WithTenantConfig(r.Context(), nil))

		defer func() {
			if err := recover(); err != nil {
				tConfig := config.GetTenantConfig(r.Context())
				var logHook logrus.Hook
				if tConfig == nil {
					logHook = logging.NewDefaultLogHook(nil)
				} else {
					logHook = logging.NewDefaultLogHook(tConfig.DefaultSensitiveLoggerValues())
				}
				loggerFactory := logging.NewFactoryFromRequest(r, logHook, sentry.NewLogHookFromContext(r.Context()))
				logger := loggerFactory.NewLogger("recovery")

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
