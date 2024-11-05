package webapp

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/auth/handler/webapp/viewmodels"
	"github.com/authgear/authgear-server/pkg/util/template"
)

var TemplateWebNotFoundHTML = template.RegisterHTML(
	"web/not_found.html",
	Components...,
)

type NotFoundHandler struct {
	ControllerFactory ControllerFactory
	BaseViewModel     *viewmodels.BaseViewModeler
	Renderer          Renderer
}

func (h *NotFoundHandler) GetData(r *http.Request, w http.ResponseWriter) (map[string]interface{}, error) {
	data := make(map[string]interface{})
	baseViewModel := h.BaseViewModel.ViewModel(r, w)
	viewmodels.Embed(data, baseViewModel)
	return data, nil
}

func (h *NotFoundHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	data, err := h.GetData(r, w)
	if err != nil {
		panic(err)
	}

	h.Renderer.RenderHTMLStatus(w, r, http.StatusNotFound, TemplateWebNotFoundHTML, data)
}
