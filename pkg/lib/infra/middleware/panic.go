package middleware

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/authgear/authgear-server/pkg/api"
	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/util/errorutil"
	"github.com/authgear/authgear-server/pkg/util/log"
)

type PanicMiddlewareLogger struct{ *log.Logger }

func NewPanicMiddlewareLogger(lf *log.Factory) PanicMiddlewareLogger {
	return PanicMiddlewareLogger{lf.New("panic-middleware")}
}

type PanicMiddleware struct {
	Logger PanicMiddlewareLogger
}

func (m *PanicMiddleware) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				// Make error.
				var e error
				if ee, isErr := err.(error); isErr {
					e = ee
				} else {
					e = fmt.Errorf("%+v", err)
				}

				// Do not log context canceled error.
				if !errors.Is(e, context.Canceled) {
					m.Logger.WithError(e).
						WithField("stack", errorutil.Callers(10000)).
						Error("panic occurred")
				}

				// Write the error as JSON.
				// Note this will not always be successful,
				// because the downstream may have written the response.
				// In that case, this following has no effect,
				// and will generate a warning saying
				// `http: superfluous response.WriteHeader call from ...`
				apiError := apierrors.AsAPIError(e)
				resp := &api.Response{Error: e}
				httpStatus := apiError.Code
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(httpStatus)
				encoder := json.NewEncoder(w)
				_ = encoder.Encode(resp)
			}
		}()

		next.ServeHTTP(w, r)
	})
}
