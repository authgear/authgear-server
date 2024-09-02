package webapp

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/felixge/httpsnoop"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/auth/handler/webapp/viewmodels"
	"github.com/authgear/authgear-server/pkg/auth/webapp"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/log"
	"github.com/authgear/authgear-server/pkg/util/panicutil"
	"github.com/authgear/authgear-server/pkg/util/template"
)

type PanicMiddlewareLogger struct{ *log.Logger }

func NewPanicMiddlewareLogger(lf *log.Factory) PanicMiddlewareLogger {
	return PanicMiddlewareLogger{lf.New("webapp-panic-middleware")}
}

type PanicMiddlewareEndpointsProvider interface {
	ErrorEndpointURL() *url.URL
}

type PanicMiddlewareUIImplementationService interface {
	GetUIImplementation() config.UIImplementation
}

type PanicMiddleware struct {
	ErrorService            *webapp.ErrorService
	Logger                  PanicMiddlewareLogger
	BaseViewModel           *viewmodels.BaseViewModeler
	Renderer                Renderer
	Endpoints               PanicMiddlewareEndpointsProvider
	UIImplementationService PanicMiddlewareUIImplementationService
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
				cookie, cookieErr := m.ErrorService.SetRecoverableError(r, apiError)
				if cookieErr != nil {
					panic(cookieErr)
				}
				uiImpl := m.UIImplementationService.GetUIImplementation()
				r.AddCookie(cookie)

				if !written {
					switch r.Method {
					case "GET":
						fallthrough
					case "HEAD":
						// Render the HTML directly and DO NOT redirect.
						// If we redirect to the original URL, then GET request will result in infinite redirect.
						// See https://github.com/authgear/authgear-server/issues/3509

						data := make(map[string]interface{})
						baseViewModel := m.BaseViewModel.ViewModel(r, w)
						viewmodels.Embed(data, baseViewModel)
						var errorHTML *template.HTML
						switch uiImpl {
						case config.UIImplementationAuthflowV2:
							errorHTML = TemplateV2WebFatalErrorHTML
						case config.UIImplementationInteraction:
							fallthrough
						case config.UIImplementationAuthflow:
							errorHTML = TemplateWebFatalErrorHTML
						default:
							panic(fmt.Errorf("unexpected ui implementation %s", uiImpl))
						}

						m.Renderer.RenderHTML(w, r, errorHTML, data)
					default:
						r.URL.Path = m.Endpoints.ErrorEndpointURL().Path
						result := &webapp.Result{
							// Show the error in the original page.
							// The panic may come from an I/O error, which could recover by retrying.
							RedirectURI:      r.URL.String(),
							NavigationAction: "replace",
						}
						result.WriteResponse(w, r)
					}
				}
			}
		}()
		next.ServeHTTP(w, r)
	})
}
