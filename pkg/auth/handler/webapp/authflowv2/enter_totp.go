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

var TemplateWebAuthflowEnterTOTPHTML = template.RegisterHTML(
	"web/authflowv2/enter_totp.html",
	handlerwebapp.Components...,
)

var AuthflowEnterTOTPSchema = validation.NewSimpleSchema(`
	{
		"type": "object",
		"properties": {
			"x_code": { "type": "string" }
		},
		"required": ["x_code"]
	}
`)

func ConfigureAuthflowV2EnterTOTPRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "POST", "GET").
		WithPathPattern(AuthflowV2RouteEnterTOTP)
}

type AuthflowV2EnterTOTPViewModel struct {
	IsBotProtectionRequired bool
}

type AuthflowV2EnterTOTPHandler struct {
	Controller                             *handlerwebapp.AuthflowController
	BaseViewModel                          *viewmodels.BaseViewModeler
	InlinePreviewAuthflowBranchViewModeler *viewmodels.InlinePreviewAuthflowBranchViewModeler
	Renderer                               handlerwebapp.Renderer
}

func NewAuthflowV2EnterTOTPViewModel(s *webapp.Session, screen *webapp.AuthflowScreenWithFlowResponse) AuthflowV2EnterTOTPViewModel {
	// Ignore error, bpRequire would be false
	bpRequired, _ := webapp.IsAuthenticateStepBotProtectionRequired(config.AuthenticationFlowAuthenticationSecondaryTOTP, screen.StateTokenFlowResponse)

	return AuthflowV2EnterTOTPViewModel{
		IsBotProtectionRequired: bpRequired,
	}
}

func (h *AuthflowV2EnterTOTPHandler) GetData(w http.ResponseWriter, r *http.Request, s *webapp.Session, screen *webapp.AuthflowScreenWithFlowResponse) (map[string]interface{}, error) {
	data := make(map[string]interface{})

	screenViewModel := NewAuthflowV2EnterTOTPViewModel(s, screen)
	viewmodels.Embed(data, screenViewModel)

	baseViewModel := h.BaseViewModel.ViewModelForAuthFlow(r, w)
	viewmodels.Embed(data, baseViewModel)

	branchViewModel := viewmodels.NewAuthflowBranchViewModel(screen)
	viewmodels.Embed(data, branchViewModel)

	return data, nil
}

func (h *AuthflowV2EnterTOTPHandler) GetInlinePreviewData(w http.ResponseWriter, r *http.Request) (map[string]interface{}, error) {
	data := make(map[string]interface{})

	baseViewModel := h.BaseViewModel.ViewModelForInlinePreviewAuthFlow(r, w)
	viewmodels.Embed(data, baseViewModel)

	branchViewModel := h.InlinePreviewAuthflowBranchViewModeler.NewAuthflowBranchViewModelForInlinePreviewEnterTOTP()
	viewmodels.Embed(data, branchViewModel)

	return data, nil
}

func (h *AuthflowV2EnterTOTPHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var handlers handlerwebapp.AuthflowControllerHandlers
	handlers.Get(func(s *webapp.Session, screen *webapp.AuthflowScreenWithFlowResponse) error {
		data, err := h.GetData(w, r, s, screen)
		if err != nil {
			return err
		}

		h.Renderer.RenderHTML(w, r, TemplateWebAuthflowEnterTOTPHTML, data)
		return nil
	})
	handlers.PostAction("submit", func(s *webapp.Session, screen *webapp.AuthflowScreenWithFlowResponse) error {
		err := AuthflowEnterTOTPSchema.Validator().ValidateValue(handlerwebapp.FormToJSON(r.Form))
		if err != nil {
			return err
		}

		code := r.Form.Get("x_code")
		requestDeviceToken := r.Form.Get("x_device_token") == "true"

		input := map[string]interface{}{
			"authentication":       config.AuthenticationFlowAuthenticationSecondaryTOTP,
			"code":                 code,
			"request_device_token": requestDeviceToken,
		}

		err = handlerwebapp.HandleAuthenticationBotProtection(config.AuthenticationFlowAuthenticationSecondaryTOTP, screen.StateTokenFlowResponse, r.Form, input)
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
	handlers.InlinePreview(func(w http.ResponseWriter, r *http.Request) error {
		data, err := h.GetInlinePreviewData(w, r)
		if err != nil {
			return err
		}
		h.Renderer.RenderHTML(w, r, TemplateWebAuthflowEnterTOTPHTML, data)
		return nil
	})

	h.Controller.HandleStep(w, r, &handlers)
}
