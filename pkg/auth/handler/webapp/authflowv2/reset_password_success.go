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

var TemplateWebAuthflowResetPasswordSuccessHTML = template.RegisterHTML(
	"web/authflowv2/reset_password_success.html",
	handlerwebapp.Components...,
)

func ConfigureAuthflowV2ResetPasswordSuccessRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "GET").
		WithPathPattern(AuthflowV2RouteResetPasswordSuccess)
}

type AuthflowV2ResetPasswordSuccessHandler struct {
	Controller    *handlerwebapp.AuthflowController
	BaseViewModel *viewmodels.BaseViewModeler
	Renderer      handlerwebapp.Renderer
}

type AuthflowV2ResetPasswordSuccessViewModel struct {
	CanBackToSignIn bool
}

func (h *AuthflowV2ResetPasswordSuccessHandler) GetData(w http.ResponseWriter, r *http.Request) (map[string]interface{}, error) {
	data := make(map[string]interface{})

	baseViewModel := h.BaseViewModel.ViewModelForAuthFlow(r, w)
	resetPasswordSuccessViewModel := &AuthflowV2ResetPasswordSuccessViewModel{
		CanBackToSignIn: r.URL.Query().Get("x_can_back_to_login") == "true",
	}
	viewmodels.Embed(data, baseViewModel)
	viewmodels.Embed(data, resetPasswordSuccessViewModel)

	return data, nil
}

func (h *AuthflowV2ResetPasswordSuccessHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var handlers handlerwebapp.AuthflowControllerHandlers
	handlers.Get(func(ctx context.Context, s *webapp.Session, _ *webapp.AuthflowScreenWithFlowResponse) error {
		data, err := h.GetData(w, r)
		if err != nil {
			return err
		}

		h.Renderer.RenderHTML(w, r, TemplateWebAuthflowResetPasswordSuccessHTML, data)
		return nil
	})

	h.Controller.HandleWithoutSession(r.Context(), w, r, &handlers)
}
