package authflowv2

import (
	"net/http"

	handlerwebapp "github.com/authgear/authgear-server/pkg/auth/handler/webapp"
	"github.com/authgear/authgear-server/pkg/auth/handler/webapp/viewmodels"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/template"
)

func ConfigureNoProjectPreviewWidgetRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("GET").
		WithPathPattern("/noproject/preview/widget")
}

var TemplateWebAuthflowPreviewWidgetHTML = template.RegisterHTML(
	"web/authflowv2/preview_widget.html",
	handlerwebapp.Components...,
)

type PreviewWidgetHandler struct {
	BaseViewModeler *viewmodels.NoProjectBaseViewModeler
	Renderer        handlerwebapp.Renderer
}

func (h *PreviewWidgetHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	data := map[string]interface{}{}

	baseViewModel := h.BaseViewModeler.ViewModel(r, w)
	viewmodels.Embed(data, baseViewModel)
	h.Renderer.RenderHTML(w, r, TemplateWebAuthflowPreviewWidgetHTML, data)
}
