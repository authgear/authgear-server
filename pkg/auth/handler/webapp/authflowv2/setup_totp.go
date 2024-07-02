package authflowv2

import (
	htmltemplate "html/template"
	"net/http"

	handlerwebapp "github.com/authgear/authgear-server/pkg/auth/handler/webapp"
	"github.com/authgear/authgear-server/pkg/auth/handler/webapp/viewmodels"
	"github.com/authgear/authgear-server/pkg/auth/webapp"
	"github.com/authgear/authgear-server/pkg/lib/authenticationflow/declarative"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	coreimage "github.com/authgear/authgear-server/pkg/util/image"
	"github.com/authgear/authgear-server/pkg/util/secretcode"
	"github.com/authgear/authgear-server/pkg/util/template"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

var TemplateWebAuthflowSetupTOTPHTML = template.RegisterHTML(
	"web/authflowv2/setup_totp.html",
	handlerwebapp.Components...,
)

var TemplateWebAuthflowSetupTOTPVerifyHTML = template.RegisterHTML(
	"web/authflowv2/setup_totp_verify.html",
	handlerwebapp.Components...,
)

var AuthflowSetupTOTPSchema = validation.NewSimpleSchema(`
	{
		"type": "object",
		"properties": {
			"x_code": { "type": "string" }
		},
		"required": ["x_code"]
	}
`)

func ConfigureAuthflowV2SetupTOTPRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "POST", "GET").
		WithPathPattern(AuthflowV2RouteSetupTOTP)
}

type AuthflowSetupTOTPViewModel struct {
	Secret   string
	ImageURI htmltemplate.URL
}

type AuthflowV2SetupTOTPHandler struct {
	Controller    *handlerwebapp.AuthflowController
	BaseViewModel *viewmodels.BaseViewModeler
	Renderer      handlerwebapp.Renderer
}

func (h *AuthflowV2SetupTOTPHandler) GetData(w http.ResponseWriter, r *http.Request, s *webapp.Session, screen *webapp.AuthflowScreenWithFlowResponse) (map[string]interface{}, error) {
	data := make(map[string]interface{})

	baseViewModel := h.BaseViewModel.ViewModelForAuthFlow(r, w)
	viewmodels.Embed(data, baseViewModel)

	screenData := screen.StateTokenFlowResponse.Action.Data.(declarative.IntentCreateAuthenticatorTOTPData)

	img, err := secretcode.QRCodeImageFromURI(screenData.OTPAuthURI, 512, 512)
	if err != nil {
		return nil, err
	}
	dataURI, err := coreimage.DataURIFromImage(coreimage.CodecPNG, img)
	if err != nil {
		return nil, err
	}

	screenViewModel := AuthflowSetupTOTPViewModel{
		Secret: screenData.Secret,
		// nolint: gosec
		ImageURI: htmltemplate.URL(dataURI),
	}
	viewmodels.Embed(data, screenViewModel)

	branchViewModel := viewmodels.NewAuthflowBranchViewModel(screen)
	viewmodels.Embed(data, branchViewModel)

	return data, nil
}

func (h *AuthflowV2SetupTOTPHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var handlers handlerwebapp.AuthflowControllerHandlers
	handlers.Get(func(s *webapp.Session, screen *webapp.AuthflowScreenWithFlowResponse) error {
		data, err := h.GetData(w, r, s, screen)
		if err != nil {
			return err
		}

		if r.URL.Query().Get("q_setup_totp_step") == "verify" {
			h.Renderer.RenderHTML(w, r, TemplateWebAuthflowSetupTOTPVerifyHTML, data)
			return nil
		}

		h.Renderer.RenderHTML(w, r, TemplateWebAuthflowSetupTOTPHTML, data)
		return nil
	})
	handlers.PostAction("submit", func(s *webapp.Session, screen *webapp.AuthflowScreenWithFlowResponse) error {
		err := AuthflowSetupTOTPSchema.Validator().ValidateValue(handlerwebapp.FormToJSON(r.Form))
		if err != nil {
			return err
		}

		code := r.Form.Get("x_code")

		input := map[string]interface{}{
			"code": code,
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
