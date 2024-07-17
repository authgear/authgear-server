package authflowv2

import (
	"net/http"

	handlerwebapp "github.com/authgear/authgear-server/pkg/auth/handler/webapp"
	"github.com/authgear/authgear-server/pkg/auth/handler/webapp/viewmodels"
	"github.com/authgear/authgear-server/pkg/auth/webapp"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/template"
)

var TemplateWebAuthflowVerifyBotProtectionHTML = template.RegisterHTML(
	"web/authflowv2/verify_bot_protection.html",
	handlerwebapp.Components...,
)

func ConfigureAuthflowV2VerifyBotProtectionRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "POST", "GET").
		WithPathPattern(AuthflowV2RouteVerifyBotProtection)
}

type AuthflowV2VerifyBotProtectionHandler struct {
	Controller    *handlerwebapp.AuthflowController
	BaseViewModel *viewmodels.BaseViewModeler
	Renderer      handlerwebapp.Renderer
}

func (h *AuthflowV2VerifyBotProtectionHandler) GetData(w http.ResponseWriter, r *http.Request, screen *webapp.AuthflowScreenWithFlowResponse) (map[string]interface{}, error) {
	data := make(map[string]interface{})
	baseViewModel := h.BaseViewModel.ViewModelForAuthFlow(r, w)
	viewmodels.Embed(data, baseViewModel)
	return data, nil
}

func (h *AuthflowV2VerifyBotProtectionHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.Renderer.RenderHTML(w, r, TemplateWebAuthflowVerifyBotProtectionHTML, map[string]interface{}{})
	// TODO: Implement GetData & POST("")
}
