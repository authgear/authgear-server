package authflowv2

import (
	"net/http"

	handlerwebapp "github.com/authgear/authgear-server/pkg/auth/handler/webapp"
	"github.com/authgear/authgear-server/pkg/auth/handler/webapp/viewmodels"
	"github.com/authgear/authgear-server/pkg/util/template"
)

var TemplateWebNotFoundHTML = template.RegisterHTML(
	"web/authflowv2/not_found.html",
	handlerwebapp.Components...,
)

type AuthflowV2NotFoundHandler struct {
	ControllerFactory handlerwebapp.ControllerFactory
	BaseViewModel     *viewmodels.BaseViewModeler
	Renderer          handlerwebapp.Renderer
}

func (h *AuthflowV2NotFoundHandler) GetData(r *http.Request, w http.ResponseWriter) (map[string]interface{}, error) {
	data := make(map[string]interface{})
	baseViewModel := h.BaseViewModel.ViewModel(r, w)
	viewmodels.Embed(data, baseViewModel)
	return data, nil
}

func (h *AuthflowV2NotFoundHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	data, err := h.GetData(r, w)
	if err != nil {
		panic(err)
	}

	h.Renderer.RenderHTMLStatus(w, r, http.StatusNotFound, TemplateWebNotFoundHTML, data)
}
