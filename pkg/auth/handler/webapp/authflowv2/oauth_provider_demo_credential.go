package authflowv2

import (
	"net/http"

	handlerwebapp "github.com/authgear/authgear-server/pkg/auth/handler/webapp"
	"github.com/authgear/authgear-server/pkg/auth/handler/webapp/viewmodels"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/template"
)

var TemplateWebAuthflowV2OAuthProviderDemoCredentialHTML = template.RegisterHTML(
	"web/authflowv2/oauth_provider_demo_credential.html",
	handlerwebapp.Components...,
)

func ConfigureAuthflowV2OAuthProviderDemoCredentialRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "GET").
		WithPathPattern(AuthflowV2RouteOAuthProviderDemoCredential)
}

type AuthflowV2OAuthProviderDemoCredentialHandler struct {
	BaseViewModel *viewmodels.BaseViewModeler
	Renderer      handlerwebapp.Renderer
}

func (h *AuthflowV2OAuthProviderDemoCredentialHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	data := make(map[string]interface{})

	baseViewModel := h.BaseViewModel.ViewModelForAuthFlow(r, w)
	viewmodels.Embed(data, baseViewModel)

	h.Renderer.RenderHTML(w, r, TemplateWebAuthflowV2OAuthProviderDemoCredentialHTML, data)
}
