package authflowv2

import (
	"net/http"

	handlerwebapp "github.com/authgear/authgear-server/pkg/auth/handler/webapp"
	"github.com/authgear/authgear-server/pkg/auth/handler/webapp/viewmodels"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/template"
)

var TemplateWebAuthflowV2OAuthProviderMissingCredentialsHTML = template.RegisterHTML(
	"web/authflowv2/oauth_provider_missing_credential.html",
	handlerwebapp.Components...,
)

func ConfigureAuthflowV2OAuthProviderMissingCredentialsRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "GET").
		WithPathPattern(AuthflowV2RouteOAuthProviderMissingCredentials)
}

type AuthflowV2OAuthProviderMissingCredentialsHandler struct {
	BaseViewModel *viewmodels.BaseViewModeler
	Renderer      handlerwebapp.Renderer
}

func (h *AuthflowV2OAuthProviderMissingCredentialsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	data := make(map[string]interface{})

	baseViewModel := h.BaseViewModel.ViewModelForAuthFlow(r, w)
	viewmodels.Embed(data, baseViewModel)

	h.Renderer.RenderHTML(w, r, TemplateWebAuthflowV2OAuthProviderMissingCredentialsHTML, data)
}
