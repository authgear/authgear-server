package middleware

import (
	"log/slog"
	"net/http"

	"github.com/felixge/httpsnoop"

	"github.com/authgear/authgear-server/pkg/api"
	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/util/panicutil"
	"github.com/authgear/authgear-server/pkg/util/slogutil"
)

var PanicMiddlewareLogger = slogutil.NewLogger("panic-middleware")

type PanicMiddleware struct{}

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
				ctx := r.Context()
				logger := PanicMiddlewareLogger.GetLogger(ctx)
				logger = logger.With(slog.Bool("written", written))
				logger.WithError(e).Error(ctx, "panic occurred")

				// Write the error as JSON.
				if !written {
					apiError := apierrors.AsAPIErrorWithContext(ctx, e)
					resp := &api.Response{Error: e}
					httpStatus := apiError.Code
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(httpStatus)
					bodyBytes, err := resp.EncodeToJSON(ctx)
					if err != nil {
						// If encoding the error response fails, there's not much else we can do.
						// Log the error and proceed without writing a body.
						logger.WithError(err).Error(ctx, "failed to encode error response")
						return
					}
					_, _ = w.Write(bodyBytes)
				}
			}
		}()

		next.ServeHTTP(w, r)
	})
}
