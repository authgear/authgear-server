package authflowv2

import (
	"context"
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
	IsBotProtectionRequired bool
}

type AuthflowV2EnterPasswordHandler struct {
	Controller                             *handlerwebapp.AuthflowController
	BaseViewModel                          *viewmodels.BaseViewModeler
	Renderer                               handlerwebapp.Renderer
	InlinePreviewAuthflowBranchViewModeler *viewmodels.InlinePreviewAuthflowBranchViewModeler
}

func NewAuthflowEnterPasswordViewModel(ctx context.Context, s *webapp.Session, screen *webapp.AuthflowScreenWithFlowResponse) AuthflowEnterPasswordViewModel {
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

		if err := phoneFormat.CheckFormat(ctx, loginID); err == nil {
			forgotPasswordInputType = "phone"
			forgotPasswordLoginID = loginID
		} else if err := emailFormat.CheckFormat(ctx, loginID); err == nil {
			forgotPasswordInputType = "email"
			forgotPasswordLoginID = loginID
		}
	}

	// Ignore error, bpRequire would be false
	bpRequired, _ := webapp.IsAuthenticateStepBotProtectionRequired(option.Authentication, screen.StateTokenFlowResponse)

	return AuthflowEnterPasswordViewModel{
		AuthenticationStage:     string(authenticationStage),
		PasswordManagerUsername: passwordManagerUsername,
		ForgotPasswordInputType: forgotPasswordInputType,
		ForgotPasswordLoginID:   forgotPasswordLoginID,
		IsBotProtectionRequired: bpRequired,
	}
}

func NewInlinePreviewAuthflowEnterPasswordViewModel() AuthflowEnterPasswordViewModel {
	return AuthflowEnterPasswordViewModel{
		AuthenticationStage:     string(authn.AuthenticationStagePrimary),
		PasswordManagerUsername: "",
		ForgotPasswordInputType: "",
		ForgotPasswordLoginID:   "",
	}
}

func (h *AuthflowV2EnterPasswordHandler) GetData(ctx context.Context, w http.ResponseWriter, r *http.Request, s *webapp.Session, screen *webapp.AuthflowScreenWithFlowResponse) (map[string]interface{}, error) {
	data := make(map[string]interface{})

	baseViewModel := h.BaseViewModel.ViewModelForAuthFlow(r, w)
	viewmodels.Embed(data, baseViewModel)

	screenViewModel := NewAuthflowEnterPasswordViewModel(ctx, s, screen)
	viewmodels.Embed(data, screenViewModel)

	branchViewModel := viewmodels.NewAuthflowBranchViewModel(screen)
	viewmodels.Embed(data, branchViewModel)

	return data, nil
}

func (h *AuthflowV2EnterPasswordHandler) GetInlinePreviewData(w http.ResponseWriter, r *http.Request) (map[string]interface{}, error) {
	data := make(map[string]interface{})

	baseViewModel := h.BaseViewModel.ViewModelForInlinePreviewAuthFlow(r, w)
	viewmodels.Embed(data, baseViewModel)

	screenViewModel := NewInlinePreviewAuthflowEnterPasswordViewModel()
	viewmodels.Embed(data, screenViewModel)

	branchViewModel := h.InlinePreviewAuthflowBranchViewModeler.NewAuthflowBranchViewModelForInlinePreviewEnterPassword()
	viewmodels.Embed(data, branchViewModel)

	return data, nil
}

func (h *AuthflowV2EnterPasswordHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var handlers handlerwebapp.AuthflowControllerHandlers
	handlers.Get(func(ctx context.Context, s *webapp.Session, screen *webapp.AuthflowScreenWithFlowResponse) error {
		data, err := h.GetData(ctx, w, r, s, screen)
		if err != nil {
			return err
		}

		h.Renderer.RenderHTML(w, r, TemplateWebAuthflowEnterPasswordHTML, data)
		return nil
	})
	handlers.PostAction("", func(ctx context.Context, s *webapp.Session, screen *webapp.AuthflowScreenWithFlowResponse) error {
		err := AuthflowEnterPasswordSchema.Validator().ValidateValue(ctx, handlerwebapp.FormToJSON(r.Form))
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

		err = handlerwebapp.HandleAuthenticationBotProtection(ctx, option.Authentication, screen.StateTokenFlowResponse, r.Form, input)
		if err != nil {
			return err
		}

		result, err := h.Controller.AdvanceWithInput(ctx, r, s, screen, input, nil)
		if err != nil {
			return err
		}

		result.WriteResponse(w, r)
		return nil
	})
	handlers.InlinePreview(func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
		data, err := h.GetInlinePreviewData(w, r)
		if err != nil {
			return err
		}
		h.Renderer.RenderHTML(w, r, TemplateWebAuthflowEnterPasswordHTML, data)
		return nil
	})

	h.Controller.HandleStep(r.Context(), w, r, &handlers)
}
