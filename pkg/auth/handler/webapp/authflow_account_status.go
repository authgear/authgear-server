package webapp

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/auth/handler/webapp/viewmodels"
	"github.com/authgear/authgear-server/pkg/auth/webapp"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/template"
)

var TemplateWebAuthflowAccountStatusHTML = template.RegisterHTML(
	"web/authflow_account_status.html",
	Components...,
)

func ConfigureAuthflowAccountStatusRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "GET").
		WithPathPattern(webapp.AuthflowRouteAccountStatus)
}

type AuthflowAccountStatusHandler struct {
	BaseViewModel *viewmodels.BaseViewModeler
	Renderer      Renderer
}

func (h *AuthflowAccountStatusHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	data := make(map[string]interface{})

	baseViewModel := h.BaseViewModel.ViewModelForAuthFlow(r, w)
	viewmodels.Embed(data, baseViewModel)

	h.Renderer.RenderHTML(w, r, TemplateWebAuthflowAccountStatusHTML, data)
}
