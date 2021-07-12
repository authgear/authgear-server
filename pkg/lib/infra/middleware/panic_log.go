package middleware

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/authgear/authgear-server/pkg/util/errorutil"
	"github.com/authgear/authgear-server/pkg/util/log"
)

type PanicLogMiddlewareLogger struct{ *log.Logger }

func NewPanicLogMiddlewareLogger(lf *log.Factory) PanicLogMiddlewareLogger {
	return PanicLogMiddlewareLogger{lf.New("log-panic")}
}

type PanicLogMiddleware struct {
	Logger PanicLogMiddlewareLogger
}

func (m *PanicLogMiddleware) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				var e error
				if ee, isErr := err.(error); isErr {
					e = ee
				} else {
					e = fmt.Errorf("%+v", err)
				}

				// ignore context cancel error
				if !errors.Is(e, context.Canceled) {
					m.Logger.WithError(e).
						WithField("stack", errorutil.Callers(10000)).
						Error("panic occurred")
				}

				// Rethrow
				panic(err)
			}
		}()

		next.ServeHTTP(w, r)
	})
}
