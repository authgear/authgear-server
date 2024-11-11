package authflowv2

import (
	"context"
	"net/http"

	handlerwebapp "github.com/authgear/authgear-server/pkg/auth/handler/webapp"
	"github.com/authgear/authgear-server/pkg/auth/handler/webapp/viewmodels"
	"github.com/authgear/authgear-server/pkg/auth/webapp"
	"github.com/authgear/authgear-server/pkg/lib/authenticationflow/declarative"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/template"
)

var TemplateWebAuthflowViewRecoveryCodeHTML = template.RegisterHTML(
	"web/authflowv2/view_recovery_code.html",
	handlerwebapp.Components...,
)

func ConfigureAuthflowV2ViewRecoveryCodeRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "POST", "GET").
		WithPathPattern(AuthflowV2RouteViewRecoveryCode)
}

type AuthflowViewRecoveryCodeViewModel struct {
	RecoveryCodes []string
}

type AuthflowV2ViewRecoveryCodeHandler struct {
	Controller    *handlerwebapp.AuthflowController
	BaseViewModel *viewmodels.BaseViewModeler
	Renderer      handlerwebapp.Renderer
}

func (h *AuthflowV2ViewRecoveryCodeHandler) GetData(w http.ResponseWriter, r *http.Request, s *webapp.Session, screen *webapp.AuthflowScreenWithFlowResponse) (map[string]interface{}, error) {
	data := map[string]interface{}{}

	baseViewModel := h.BaseViewModel.ViewModelForAuthFlow(r, w)
	viewmodels.Embed(data, baseViewModel)

	screenData := screen.StateTokenFlowResponse.Action.Data.(declarative.IntentSignupFlowStepViewRecoveryCodeData)

	screenViewModel := AuthflowViewRecoveryCodeViewModel{
		RecoveryCodes: handlerwebapp.FormatRecoveryCodes(screenData.RecoveryCodes),
	}
	viewmodels.Embed(data, screenViewModel)

	return data, nil
}

func (h *AuthflowV2ViewRecoveryCodeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var handlers handlerwebapp.AuthflowControllerHandlers
	handlers.Get(func(ctx context.Context, s *webapp.Session, screen *webapp.AuthflowScreenWithFlowResponse) error {
		data, err := h.GetData(w, r, s, screen)
		if err != nil {
			return err
		}

		h.Renderer.RenderHTML(w, r, TemplateWebAuthflowViewRecoveryCodeHTML, data)
		return nil
	})
	handlers.PostAction("download", func(ctx context.Context, s *webapp.Session, screen *webapp.AuthflowScreenWithFlowResponse) error {
		data, err := h.GetData(w, r, s, screen)
		if err != nil {
			return err
		}

		handlerwebapp.SetRecoveryCodeAttachmentHeaders(w)
		h.Renderer.Render(w, r, handlerwebapp.TemplateWebDownloadRecoveryCodeTXT, data)
		return nil
	})
	handlers.PostAction("proceed", func(ctx context.Context, s *webapp.Session, screen *webapp.AuthflowScreenWithFlowResponse) error {
		input := map[string]interface{}{
			"confirm_recovery_code": true,
		}

		result, err := h.Controller.AdvanceWithInput(ctx, r, s, screen, input, nil)
		if err != nil {
			return err
		}

		result.WriteResponse(w, r)
		return nil
	})
	h.Controller.HandleStep(r.Context(), w, r, &handlers)
}
