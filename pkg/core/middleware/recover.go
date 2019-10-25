package middleware

import (
	"net/http"

	"github.com/skygeario/skygear-server/pkg/core/errors"
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

				var e error
				if ee, isErr := err.(error); isErr {
					e = ee
				} else {
					e = errors.Newf("%+v", err)
				}
				logger.WithError(e).WithField("stack", errors.Callers(8)).Error("unexpected panic occurred")

				w.WriteHeader(http.StatusInternalServerError)
			}
		}()

		next.ServeHTTP(w, r)
	})
}
