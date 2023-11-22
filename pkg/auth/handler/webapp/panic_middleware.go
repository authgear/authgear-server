package webapp

import (
	"net/http"

	"github.com/felixge/httpsnoop"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/auth/handler/webapp/viewmodels"
	"github.com/authgear/authgear-server/pkg/auth/webapp"
	"github.com/authgear/authgear-server/pkg/util/log"
	"github.com/authgear/authgear-server/pkg/util/panicutil"
)

type PanicMiddlewareLogger struct{ *log.Logger }

func NewPanicMiddlewareLogger(lf *log.Factory) PanicMiddlewareLogger {
	return PanicMiddlewareLogger{lf.New("webapp-panic-middleware")}
}

type PanicMiddleware struct {
	ErrorCookie   *webapp.ErrorCookie
	Logger        PanicMiddlewareLogger
	BaseViewModel *viewmodels.BaseViewModeler
	Renderer      Renderer
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
			if e := recover(); e != nil {
				err := panicutil.MakeError(e)
				m.Logger.WithError(err).Error("panic occurred")

				apiError := apierrors.AsAPIError(err)

				if !written {
					cookie, cookieErr := m.ErrorCookie.SetRecoverableError(r, apiError)
					if cookieErr != nil {
						panic(cookieErr)
					}

					result := &webapp.Result{
						// Show the error in the original page.
						// The panic may come from an I/O error, which could recover by retrying.
						RedirectURI:      r.URL.String(),
						NavigationAction: "replace",
						Cookies:          []*http.Cookie{cookie},
					}
					result.WriteResponse(w, r)
				}
			}
		}()
		next.ServeHTTP(w, r)
	})
}
