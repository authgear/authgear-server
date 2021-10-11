package webapp

import (
	"net/http"

	"github.com/felixge/httpsnoop"

	"github.com/authgear/authgear-server/pkg/auth/handler/webapp/viewmodels"
	"github.com/authgear/authgear-server/pkg/util/log"
	"github.com/authgear/authgear-server/pkg/util/panicutil"
	"github.com/authgear/authgear-server/pkg/util/template"
)

var TemplateWebFatalErrorHTML = template.RegisterHTML(
	"web/fatal_error.html",
	components...,
)

type PanicMiddlewareLogger struct{ *log.Logger }

func NewPanicMiddlewareLogger(lf *log.Factory) PanicMiddlewareLogger {
	return PanicMiddlewareLogger{lf.New("webapp-panic-middleware")}
}

type PanicMiddleware struct {
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
				log.PanicValue(m.Logger.Logger, err)

				if !written {
					data := make(map[string]interface{})
					baseViewModel := m.BaseViewModel.ViewModel(r, w)
					baseViewModel.SetError(err)
					viewmodels.Embed(data, baseViewModel)
					m.Renderer.RenderHTML(w, r, TemplateWebFatalErrorHTML, data)
				}
			}
		}()
		next.ServeHTTP(w, r)
	})
}
