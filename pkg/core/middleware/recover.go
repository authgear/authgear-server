package middleware

import (
	"net/http"

	"github.com/skygeario/skygear-server/pkg/core/errors"
	"github.com/skygeario/skygear-server/pkg/core/handler"
	"github.com/skygeario/skygear-server/pkg/core/logging"
)

// RecoverMiddleware recover from panic
type RecoverMiddleware struct {
}

func (m RecoverMiddleware) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				// TODO: read from tconfig
				logHook := logging.NewDefaultLogHook(nil)
				loggerFactory := logging.NewFactoryFromRequest(r, logHook)
				logger := loggerFactory.NewLogger("recovery")

				handled := false
				if herr, ok := err.(handler.HandledError); ok {
					handled = true
					err = herr.Error
				}

				var e error
				if ee, isErr := err.(error); isErr {
					e = ee
				} else {
					e = errors.Newf("%+v", err)
				}

				if handled {
					logger.WithError(e).Error("unexpected error occurred")
				} else {
					logger.WithError(e).WithField("stack", errors.Callers(8)).Error("panic occurred")
					w.WriteHeader(http.StatusInternalServerError)
				}
			}
		}()

		next.ServeHTTP(w, r)
	})
}
