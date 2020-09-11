package middleware

import (
	"fmt"
	"net/http"

	"github.com/authgear/authgear-server/pkg/lib/infra/db"
	"github.com/authgear/authgear-server/pkg/util/errorutil"
	"github.com/authgear/authgear-server/pkg/util/log"
)

type RecoveryLogger struct{ *log.Logger }

func NewRecoveryLogger(lf *log.Factory) RecoveryLogger { return RecoveryLogger{lf.New("recovery")} }

// RecoverMiddleware recover from panic
type RecoverMiddleware struct {
	Logger RecoveryLogger
}

func (m *RecoverMiddleware) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				const errorTypeUnexpected = 0
				const errorTypeConflict = 2
				errorType := errorTypeUnexpected
				if err, ok := err.(error); ok && errorutil.Is(err, db.ErrWriteConflict) {
					errorType = errorTypeConflict
				}

				var e error
				if ee, isErr := err.(error); isErr {
					e = ee
				} else {
					e = fmt.Errorf("%+v", err)
				}

				if errorType == errorTypeUnexpected {
					m.Logger.WithError(e).
						WithField("stack", errorutil.Callers(8)).
						Error("panic occurred")
					w.WriteHeader(http.StatusInternalServerError)
				} else if errorType == errorTypeConflict {
					w.WriteHeader(http.StatusConflict)
				}
			}
		}()

		next.ServeHTTP(w, r)
	})
}
