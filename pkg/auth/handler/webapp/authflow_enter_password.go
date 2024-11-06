package webapp

import (
	"context"
	"net/http"

	"github.com/authgear/authgear-server/pkg/auth/handler/webapp/viewmodels"
	"github.com/authgear/authgear-server/pkg/auth/webapp"
	"github.com/authgear/authgear-server/pkg/lib/authenticationflow/declarative"
	"github.com/authgear/authgear-server/pkg/lib/authn"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/template"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

var TemplateWebAuthflowEnterPasswordHTML = template.RegisterHTML(
	"web/authflow_enter_password.html",
	Components...,
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

func ConfigureAuthflowEnterPasswordRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "POST", "GET").
		WithPathPattern(webapp.AuthflowRouteEnterPassword)
}

type AuthflowEnterPasswordViewModel struct {
	AuthenticationStage     string
	PasswordManagerUsername string
	ForgotPasswordInputType string
	ForgotPasswordLoginID   string
}

type AuthflowEnterPasswordHandler struct {
	Controller    *AuthflowController
	BaseViewModel *viewmodels.BaseViewModeler
	Renderer      Renderer
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
	if loginID, ok := FindLoginIDInPreviousInput(s, screen.Screen.StateToken.XStep); ok {
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

func (h *AuthflowEnterPasswordHandler) GetData(w http.ResponseWriter, r *http.Request, s *webapp.Session, screen *webapp.AuthflowScreenWithFlowResponse) (map[string]interface{}, error) {
	data := make(map[string]interface{})

	baseViewModel := h.BaseViewModel.ViewModelForAuthFlow(r, w)
	viewmodels.Embed(data, baseViewModel)

	screenViewModel := NewAuthflowEnterPasswordViewModel(s, screen)
	viewmodels.Embed(data, screenViewModel)

	branchViewModel := viewmodels.NewAuthflowBranchViewModel(screen)
	viewmodels.Embed(data, branchViewModel)

	return data, nil
}

func (h *AuthflowEnterPasswordHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var handlers AuthflowControllerHandlers
	handlers.Get(func(ctx context.Context, s *webapp.Session, screen *webapp.AuthflowScreenWithFlowResponse) error {
		data, err := h.GetData(w, r, s, screen)
		if err != nil {
			return err
		}

		h.Renderer.RenderHTML(w, r, TemplateWebAuthflowEnterPasswordHTML, data)
		return nil
	})
	handlers.PostAction("", func(ctx context.Context, s *webapp.Session, screen *webapp.AuthflowScreenWithFlowResponse) error {
		err := AuthflowEnterPasswordSchema.Validator().ValidateValue(FormToJSON(r.Form))
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

		result, err := h.Controller.AdvanceWithInput(ctx, r, s, screen, input, nil)
		if err != nil {
			return err
		}

		result.WriteResponse(w, r)
		return nil
	})
	h.Controller.HandleStep(r.Context(), w, r, &handlers)
}
