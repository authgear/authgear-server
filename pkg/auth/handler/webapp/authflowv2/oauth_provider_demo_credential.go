package authflowv2

import (
	"context"
	"net/http"

	handlerwebapp "github.com/authgear/authgear-server/pkg/auth/handler/webapp"
	"github.com/authgear/authgear-server/pkg/auth/handler/webapp/viewmodels"
	"github.com/authgear/authgear-server/pkg/auth/webapp"
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
	Controller    *handlerwebapp.AuthflowController
	BaseViewModel *viewmodels.BaseViewModeler
	Renderer      handlerwebapp.Renderer
}

func (h *AuthflowV2OAuthProviderDemoCredentialHandler) GetData(w http.ResponseWriter, r *http.Request, screen *webapp.AuthflowScreenWithFlowResponse) (map[string]interface{}, error) {
	data := make(map[string]interface{})

	baseViewModel := h.BaseViewModel.ViewModelForAuthFlow(r, w)
	viewmodels.Embed(data, baseViewModel)
	if screen.Screen.OAuthProviderDemoCredentialViewModel != nil {
		viewmodels.Embed(data, screen.Screen.OAuthProviderDemoCredentialViewModel)
	}

	return data, nil
}

func (h *AuthflowV2OAuthProviderDemoCredentialHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var handlers handlerwebapp.AuthflowControllerHandlers
	handlers.Get(func(ctx context.Context, s *webapp.Session, screen *webapp.AuthflowScreenWithFlowResponse) error {
		data, err := h.GetData(w, r, screen)
		if err != nil {
			return err
		}

		h.Renderer.RenderHTML(w, r, TemplateWebAuthflowV2OAuthProviderDemoCredentialHTML, data)
		return nil
	})

	h.Controller.HandleStep(r.Context(), w, r, &handlers)
}
