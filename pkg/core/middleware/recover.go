package middleware

import (
	"net/http"

	"github.com/sirupsen/logrus"

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
				formatter := logging.NewDefaultMaskedTextFormatter(nil)
				loggerFactory := logging.NewFactoryFromRequest(r, formatter)
				logger := loggerFactory.NewLogger("recovery")

				var details errors.Details
				if err, isErr := err.(error); isErr {
					logger = logger.WithError(err)
					details = errors.CollectDetails(err, nil)
				} else {
					logger = logger.WithError(errors.Newf("%+v", err))
				}

				logger.
					WithFields(logrus.Fields{"stack": errors.Callers(8), "details": details}).
					Error("unexpected panic occurred")
			}
		}()

		next.ServeHTTP(w, r)
	})
}
