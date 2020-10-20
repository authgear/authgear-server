package webapp

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/auth/handler/webapp/viewmodels"
	"github.com/authgear/authgear-server/pkg/util/template"
)

var TemplateWebFatalErrorHTML = template.RegisterHTML(
	"web/fatal_error.html",
	components...,
)

type PanicMiddleware struct {
	BaseViewModel *viewmodels.BaseViewModeler
	Renderer      Renderer
}

func (m *PanicMiddleware) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				data := make(map[string]interface{})
				baseViewModel := m.BaseViewModel.ViewModel(r, err)
				viewmodels.Embed(data, baseViewModel)
				m.Renderer.RenderHTML(w, r, TemplateWebFatalErrorHTML, data)

				// Rethrow
				panic(err)
			}
		}()
		next.ServeHTTP(w, r)
	})
}
