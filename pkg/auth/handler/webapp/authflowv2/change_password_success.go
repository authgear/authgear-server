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

var TemplateWebAuthflowChangePasswordSuccessHTML = template.RegisterHTML(
	"web/authflowv2/change_password_success.html",
	handlerwebapp.Components...,
)

func ConfigureAuthflowV2ChangePasswordSuccessRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "GET", "POST").
		WithPathPattern(AuthflowV2RouteChangePasswordSuccess)
}

type AuthflowV2ChangePasswordSuccessHandler struct {
	Controller    *handlerwebapp.AuthflowController
	BaseViewModel *viewmodels.BaseViewModeler
	Renderer      handlerwebapp.Renderer
}

type AuthflowV2ChangePasswordSuccessViewModel struct {
	CanBackToSignIn bool
}

func (h *AuthflowV2ChangePasswordSuccessHandler) GetData(w http.ResponseWriter, r *http.Request) (map[string]interface{}, error) {
	data := make(map[string]interface{})

	baseViewModel := h.BaseViewModel.ViewModelForAuthFlow(r, w)
	viewmodels.Embed(data, baseViewModel)

	return data, nil
}

func (h *AuthflowV2ChangePasswordSuccessHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var handlers handlerwebapp.AuthflowControllerHandlers
	handlers.Get(func(ctx context.Context, s *webapp.Session, screen *webapp.AuthflowScreenWithFlowResponse) error {
		data, err := h.GetData(w, r)
		if err != nil {
			return err
		}

		h.Renderer.RenderHTML(w, r, TemplateWebAuthflowChangePasswordSuccessHTML, data)
		return nil
	})

	handlers.PostAction("", func(ctx context.Context, s *webapp.Session, screen *webapp.AuthflowScreenWithFlowResponse) error {
		result, err := h.Controller.AdvanceToDelayedScreen(r, s)
		if err != nil {
			return err
		}
		result.WriteResponse(w, r)
		return nil
	})

	h.Controller.HandleWithoutFlow(r.Context(), w, r, &handlers)
}
