package webapp

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/auth/handler/webapp/viewmodels"
	"github.com/authgear/authgear-server/pkg/auth/webapp"
	"github.com/authgear/authgear-server/pkg/lib/authenticationflow/declarative"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/template"
)

var TemplateWebAuthflowViewRecoveryCodeHTML = template.RegisterHTML(
	"web/authflow_view_recovery_code.html",
	components...,
)

func ConfigureAuthflowViewRecoveryCodeRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "POST", "GET").
		WithPathPattern(webapp.AuthflowRouteViewRecoveryCode)
}

type AuthflowViewRecoveryCodeViewModel struct {
	RecoveryCodes []string
}

type AuthflowViewRecoveryCodeHandler struct {
	Controller    *AuthflowController
	BaseViewModel *viewmodels.BaseViewModeler
	Renderer      Renderer
}

func (h *AuthflowViewRecoveryCodeHandler) GetData(w http.ResponseWriter, r *http.Request, s *webapp.Session, screen *webapp.AuthflowScreenWithFlowResponse) (map[string]interface{}, error) {
	data := map[string]interface{}{}

	baseViewModel := h.BaseViewModel.ViewModel(r, w)
	viewmodels.Embed(data, baseViewModel)

	screenData := screen.StateTokenFlowResponse.Action.Data.(declarative.IntentSignupFlowStepRecoveryCodeData)

	screenViewModel := AuthflowViewRecoveryCodeViewModel{
		RecoveryCodes: formatRecoveryCodes(screenData.RecoveryCodes),
	}
	viewmodels.Embed(data, screenViewModel)

	return data, nil
}

func (h *AuthflowViewRecoveryCodeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var handlers AuthflowControllerHandlers
	handlers.Get(func(s *webapp.Session, screen *webapp.AuthflowScreenWithFlowResponse) error {
		data, err := h.GetData(w, r, s, screen)
		if err != nil {
			return err
		}

		h.Renderer.RenderHTML(w, r, TemplateWebAuthflowViewRecoveryCodeHTML, data)
		return nil
	})
	handlers.PostAction("download", func(s *webapp.Session, screen *webapp.AuthflowScreenWithFlowResponse) error {
		data, err := h.GetData(w, r, s, screen)
		if err != nil {
			return err
		}

		setRecoveryCodeAttachmentHeaders(w)
		h.Renderer.Render(w, r, TemplateWebDownloadRecoveryCodeTXT, data)
		return nil
	})
	handlers.PostAction("proceed", func(s *webapp.Session, screen *webapp.AuthflowScreenWithFlowResponse) error {
		input := map[string]interface{}{
			"confirm_recovery_code": true,
		}

		result, err := h.Controller.AdvanceWithInput(r, s, screen, input)
		if err != nil {
			return err
		}

		result.WriteResponse(w, r)
		return nil
	})
	h.Controller.HandleStep(w, r, &handlers)
}
