package authflowv2

import (
	"net/http"

	handlerwebapp "github.com/authgear/authgear-server/pkg/auth/handler/webapp"
	"github.com/authgear/authgear-server/pkg/auth/handler/webapp/viewmodels"
	"github.com/authgear/authgear-server/pkg/auth/webapp"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/template"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

var TemplateWebAuthflowEnterRecoveryCodeHTML = template.RegisterHTML(
	"web/authflowv2/enter_recovery_code.html",
	handlerwebapp.Components...,
)

var AuthflowEnterRecoveryCodeSchema = validation.NewSimpleSchema(`
	{
		"type": "object",
		"properties": {
			"x_recovery_code": {
				"type": "string",
				"format": "x_recovery_code"
			}
		},
		"required": ["x_recovery_code"]
	}
`)

func ConfigureAuthflowV2EnterRecoveryCodeRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "POST", "GET").
		WithPathPattern(AuthflowV2RouteEnterRecoveryCode)
}

type AuthflowV2EnterRecoveryCodeViewModel struct {
	IsBotProtectionRequired bool
}

type AuthflowV2EnterRecoveryCodeHandler struct {
	Controller    *handlerwebapp.AuthflowController
	BaseViewModel *viewmodels.BaseViewModeler
	Renderer      handlerwebapp.Renderer
}

func NewAuthflowV2EnterRecoveryCodeViewModel(s *webapp.Session, screen *webapp.AuthflowScreenWithFlowResponse) AuthflowV2EnterRecoveryCodeViewModel {
	// Ignore error, bpRequire would be false
	bpRequired, _ := webapp.IsAuthenticateStepBotProtectionRequired(config.AuthenticationFlowAuthenticationRecoveryCode, screen.StateTokenFlowResponse)

	return AuthflowV2EnterRecoveryCodeViewModel{
		IsBotProtectionRequired: bpRequired,
	}
}

func (h *AuthflowV2EnterRecoveryCodeHandler) GetData(w http.ResponseWriter, r *http.Request, s *webapp.Session, screen *webapp.AuthflowScreenWithFlowResponse) (map[string]interface{}, error) {
	data := make(map[string]interface{})

	screenViewModel := NewAuthflowV2EnterRecoveryCodeViewModel(s, screen)
	viewmodels.Embed(data, screenViewModel)

	baseViewModel := h.BaseViewModel.ViewModelForAuthFlow(r, w)
	viewmodels.Embed(data, baseViewModel)

	branchViewModel := viewmodels.NewAuthflowBranchViewModel(screen)
	viewmodels.Embed(data, branchViewModel)

	return data, nil
}

func (h *AuthflowV2EnterRecoveryCodeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var handlers handlerwebapp.AuthflowControllerHandlers
	handlers.Get(func(s *webapp.Session, screen *webapp.AuthflowScreenWithFlowResponse) error {
		data, err := h.GetData(w, r, s, screen)
		if err != nil {
			return err
		}

		h.Renderer.RenderHTML(w, r, TemplateWebAuthflowEnterRecoveryCodeHTML, data)
		return nil
	})
	handlers.PostAction("", func(s *webapp.Session, screen *webapp.AuthflowScreenWithFlowResponse) error {
		err := AuthflowEnterRecoveryCodeSchema.Validator().ValidateValue(handlerwebapp.FormToJSON(r.Form))
		if err != nil {
			return err
		}

		recoveryCode := r.Form.Get("x_recovery_code")
		requestDeviceToken := r.Form.Get("x_device_token") == "true"

		input := map[string]interface{}{
			"authentication":       "recovery_code",
			"recovery_code":        recoveryCode,
			"request_device_token": requestDeviceToken,
		}

		err = handlerwebapp.HandleAuthenticationBotProtection(config.AuthenticationFlowAuthenticationRecoveryCode, screen.StateTokenFlowResponse, r.Form, input)
		if err != nil {
			return err
		}

		result, err := h.Controller.AdvanceWithInput(r, s, screen, input, nil)
		if err != nil {
			return err
		}

		result.WriteResponse(w, r)
		return nil
	})
	h.Controller.HandleStep(w, r, &handlers)
}
