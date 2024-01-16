package authflowv2

import (
	"net/http"

	handlerwebapp "github.com/authgear/authgear-server/pkg/auth/handler/webapp"
	"github.com/authgear/authgear-server/pkg/auth/handler/webapp/viewmodels"
	"github.com/authgear/authgear-server/pkg/auth/webapp"
	"github.com/authgear/authgear-server/pkg/lib/authenticationflow/declarative"
	"github.com/authgear/authgear-server/pkg/lib/authn"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/template"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

var TemplateWebAuthflowEnterPasswordHTML = template.RegisterHTML(
	"web/authflowv2/enter_password.html",
	handlerwebapp.Components...,
)

var AuthflowEnterPasswordSchema = validation.NewSimpleSchema(`
	{
		"type": "object",
		"properties": {
			"x_password": { "type": "string" }
		},
		"required": ["x_password"]
	}
`)

func ConfigureAuthflowV2EnterPasswordRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "POST", "GET").
		WithPathPattern(AuthflowV2RouteEnterPassword)
}

type AuthflowEnterPasswordViewModel struct {
	AuthenticationStage     string
	PasswordManagerUsername string
	ForgotPasswordInputType string
	ForgotPasswordLoginID   string
}

type AuthflowV2EnterPasswordHandler struct {
	Controller    *handlerwebapp.AuthflowController
	BaseViewModel *viewmodels.BaseViewModeler
	Renderer      handlerwebapp.Renderer
}

func NewAuthflowEnterPasswordViewModel(s *webapp.Session, screen *webapp.AuthflowScreenWithFlowResponse) AuthflowEnterPasswordViewModel {
	index := *screen.Screen.TakenBranchIndex
	flowResponse := screen.BranchStateTokenFlowResponse
	data := flowResponse.Action.Data.(declarative.StepAuthenticateData)
	option := data.Options[index]
	authenticationStage := authn.AuthenticationStageFromAuthenticationMethod(option.Authentication)

	// Use the previous input to derive some information.
	passwordManagerUsername := ""
	forgotPasswordInputType := ""
	forgotPasswordLoginID := ""
	if loginID, ok := handlerwebapp.FindLoginIDInPreviousInput(s, screen.Screen.StateToken.XStep); ok {
		passwordManagerUsername = loginID

		phoneFormat := validation.FormatPhone{}
		emailFormat := validation.FormatEmail{AllowName: false}

		if err := phoneFormat.CheckFormat(loginID); err == nil {
			forgotPasswordInputType = "phone"
			forgotPasswordLoginID = loginID
		} else if err := emailFormat.CheckFormat(loginID); err == nil {
			forgotPasswordInputType = "email"
			forgotPasswordLoginID = loginID
		}
	}

	return AuthflowEnterPasswordViewModel{
		AuthenticationStage:     string(authenticationStage),
		PasswordManagerUsername: passwordManagerUsername,
		ForgotPasswordInputType: forgotPasswordInputType,
		ForgotPasswordLoginID:   forgotPasswordLoginID,
	}
}

func (h *AuthflowV2EnterPasswordHandler) GetData(w http.ResponseWriter, r *http.Request, s *webapp.Session, screen *webapp.AuthflowScreenWithFlowResponse) (map[string]interface{}, error) {
	data := make(map[string]interface{})

	baseViewModel := h.BaseViewModel.ViewModelForAuthFlow(r, w)
	viewmodels.Embed(data, baseViewModel)

	screenViewModel := NewAuthflowEnterPasswordViewModel(s, screen)
	viewmodels.Embed(data, screenViewModel)

	branchViewModel := viewmodels.NewAuthflowBranchViewModel(screen)
	viewmodels.Embed(data, branchViewModel)

	return data, nil
}

func (h *AuthflowV2EnterPasswordHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var handlers handlerwebapp.AuthflowControllerHandlers
	handlers.Get(func(s *webapp.Session, screen *webapp.AuthflowScreenWithFlowResponse) error {
		data, err := h.GetData(w, r, s, screen)
		if err != nil {
			return err
		}

		h.Renderer.RenderHTML(w, r, TemplateWebAuthflowEnterPasswordHTML, data)
		return nil
	})
	handlers.PostAction("", func(s *webapp.Session, screen *webapp.AuthflowScreenWithFlowResponse) error {
		err := AuthflowEnterPasswordSchema.Validator().ValidateValue(handlerwebapp.FormToJSON(r.Form))
		if err != nil {
			return err
		}

		index := *screen.Screen.TakenBranchIndex
		flowResponse := screen.BranchStateTokenFlowResponse
		data := flowResponse.Action.Data.(declarative.StepAuthenticateData)
		option := data.Options[index]

		plainPassword := r.Form.Get("x_password")
		requestDeviceToken := r.Form.Get("x_device_token") == "true"

		input := map[string]interface{}{
			"authentication":       option.Authentication,
			"password":             plainPassword,
			"request_device_token": requestDeviceToken,
		}

		result, err := h.Controller.AdvanceWithInput(r, s, screen, input)
		if err != nil {
			return err
		}

		result.WriteResponse(w, r)
		return nil
	})

	// FIXME(newman): "resend" and "submit" are fake handlers for testing new components.
	// Remove them when the new components are ready.
	handlers.PostAction("resend", func(s *webapp.Session, screen *webapp.AuthflowScreenWithFlowResponse) error {
		input := map[string]interface{}{
			"resend": true,
		}

		result, err := h.Controller.UpdateWithInput(r, s, screen, input)
		if err != nil {
			return err
		}

		// TODO(newman): Implement flash message
		// h.FlashMessage.Flash(w, string(webapp.FlashMessageTypeResendCodeSuccess))
		result.WriteResponse(w, r)
		return nil
	})

	handlers.PostAction("submit", func(s *webapp.Session, screen *webapp.AuthflowScreenWithFlowResponse) error {
		input := map[string]interface{}{}

		result, err := h.Controller.UpdateWithInput(r, s, screen, input)
		if err != nil {
			return err
		}

		result.WriteResponse(w, r)
		return nil
	})

	h.Controller.HandleStep(w, r, &handlers)
}
