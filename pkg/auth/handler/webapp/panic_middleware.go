package webapp

import (
	"fmt"
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
			if e := recover(); e != nil {
				data := make(map[string]interface{})
				baseViewModel := m.BaseViewModel.ViewModel(r, w)

				// Make error
				var err error
				if ee, isErr := e.(error); isErr {
					err = ee
				} else {
					err = fmt.Errorf("%+v", e)
				}

				baseViewModel.SetError(err)
				viewmodels.Embed(data, baseViewModel)
				m.Renderer.RenderHTML(w, r, TemplateWebFatalErrorHTML, data)

				// Rethrow
				panic(err)
			}
		}()
		next.ServeHTTP(w, r)
	})
}
