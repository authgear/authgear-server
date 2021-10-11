package middleware

import (
	"encoding/json"
	"net/http"

	"github.com/felixge/httpsnoop"

	"github.com/authgear/authgear-server/pkg/api"
	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/util/log"
	"github.com/authgear/authgear-server/pkg/util/panicutil"
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
		written := false

		w = httpsnoop.Wrap(w, httpsnoop.Hooks{
			WriteHeader: func(f httpsnoop.WriteHeaderFunc) httpsnoop.WriteHeaderFunc {
				return func(code int) {
					written = true
					f(code)
				}
			},
			Write: func(f httpsnoop.WriteFunc) httpsnoop.WriteFunc {
				return func(b []byte) (int, error) {
					written = true
					return f(b)
				}
			},
		})

		defer func() {
			if err := recover(); err != nil {
				e := panicutil.MakeError(err)

				log.PanicValue(m.Logger.Logger, e)

				// Write the error as JSON.
				if !written {
					apiError := apierrors.AsAPIError(e)
					resp := &api.Response{Error: e}
					httpStatus := apiError.Code
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(httpStatus)
					encoder := json.NewEncoder(w)
					_ = encoder.Encode(resp)
				}
			}
		}()

		next.ServeHTTP(w, r)
	})
}
