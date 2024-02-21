package authflowv2

import (
	"net/http"

	handlerwebapp "github.com/authgear/authgear-server/pkg/auth/handler/webapp"
	"github.com/authgear/authgear-server/pkg/auth/handler/webapp/viewmodels"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/template"
)

var TemplateWebAuthflowV2NoAuthenticatorHTML = template.RegisterHTML(
	"web/authflowv2/no_authenticator.html",
	handlerwebapp.Components...,
)

func ConfigureAuthflowV2NoAuthenticatorRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "GET").
		WithPathPattern(AuthflowV2RouteNoAuthenticator)
}

type AuthflowV2NoAuthenticatorHandler struct {
	BaseViewModel *viewmodels.BaseViewModeler
	Renderer      handlerwebapp.Renderer
}

func (h *AuthflowV2NoAuthenticatorHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	data := make(map[string]interface{})

	baseViewModel := h.BaseViewModel.ViewModelForAuthFlow(r, w)
	viewmodels.Embed(data, baseViewModel)

	h.Renderer.RenderHTML(w, r, TemplateWebAuthflowV2NoAuthenticatorHTML, data)
}
