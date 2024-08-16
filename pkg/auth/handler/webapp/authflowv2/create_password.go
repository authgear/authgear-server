package authflowv2

import (
	"net/http"

	handlerwebapp "github.com/authgear/authgear-server/pkg/auth/handler/webapp"
	authflowv2viewmodels "github.com/authgear/authgear-server/pkg/auth/handler/webapp/authflowv2/viewmodels"
	"github.com/authgear/authgear-server/pkg/auth/handler/webapp/viewmodels"
	"github.com/authgear/authgear-server/pkg/auth/webapp"
	"github.com/authgear/authgear-server/pkg/lib/authenticationflow/declarative"
	"github.com/authgear/authgear-server/pkg/lib/authn"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	pwd "github.com/authgear/authgear-server/pkg/util/password"
	"github.com/authgear/authgear-server/pkg/util/template"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

var TemplateWebAuthflowCreatePasswordHTML = template.RegisterHTML(
	"web/authflowv2/create_password.html",
	handlerwebapp.Components...,
)

var AuthflowCreatePasswordSchema = validation.NewSimpleSchema(`
	{
		"type": "object",
		"properties": {
			"x_password": { "type": "string" },
			"x_confirm_password": { "type": "string" }
		},
		"required": ["x_password", "x_confirm_password"]
	}
`)

func ConfigureAuthflowV2CreatePasswordRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "POST", "GET").
		WithPathPattern(AuthflowV2RouteCreatePassword)
}

type AuthflowCreatePasswordViewModel struct {
	AuthenticationStage     string
	PasswordManagerUsername string
	ForgotPasswordInputType string
	ForgotPasswordLoginID   string
}

type AuthflowV2CreatePasswordHandler struct {
	Controller                             *handlerwebapp.AuthflowController
	BaseViewModel                          *viewmodels.BaseViewModeler
	InlinePreviewAuthflowBranchViewModeler *viewmodels.InlinePreviewAuthflowBranchViewModeler
	Renderer                               handlerwebapp.Renderer
	FeatureConfig                          *config.FeatureConfig
	AuthenticatorConfig                    *config.AuthenticatorConfig
}

func (h *AuthflowV2CreatePasswordHandler) GetData(w http.ResponseWriter, r *http.Request, s *webapp.Session, screen *webapp.AuthflowScreenWithFlowResponse) (map[string]interface{}, error) {
	data := make(map[string]interface{})

	baseViewModel := h.BaseViewModel.ViewModelForAuthFlow(r, w)
	viewmodels.Embed(data, baseViewModel)

	screenViewModel := AuthflowCreatePasswordViewModel{}

	index := *screen.Screen.TakenBranchIndex
	flowResponse := screen.BranchStateTokenFlowResponse
	screenData := flowResponse.Action.Data.(declarative.CreateAuthenticatorData)
	option := screenData.Options[index]
	authenticationStage := authn.AuthenticationStageFromAuthenticationMethod(option.Authentication)
	isPrimary := authenticationStage == authn.AuthenticationStagePrimary

	screenViewModel.AuthenticationStage = string(authenticationStage)

	if loginID, ok := handlerwebapp.FindLoginIDInPreviousInput(s, screen.Screen.StateToken.XStep); ok {
		screenViewModel.PasswordManagerUsername = loginID
	}

	passwordPolicyViewModel := viewmodels.NewPasswordPolicyViewModelFromAuthflow(
		option.PasswordPolicy,
		baseViewModel.RawError,
		&viewmodels.PasswordPolicyViewModelOptions{
			// Hide reuse password policy when creating new
			// password through web UI (sign up)
			IsNew: isPrimary,
		},
	)

	passwordInputErrorViewModel := authflowv2viewmodels.NewPasswordInputErrorViewModel(baseViewModel.RawError)

	viewmodels.Embed(data, screenViewModel)
	viewmodels.Embed(data, passwordPolicyViewModel)
	viewmodels.Embed(data, passwordInputErrorViewModel)

	branchViewModel := viewmodels.NewAuthflowBranchViewModel(screen)
	viewmodels.Embed(data, branchViewModel)

	return data, nil
}

func (h *AuthflowV2CreatePasswordHandler) GetInlinePreviewData(w http.ResponseWriter, r *http.Request) (map[string]interface{}, error) {
	data := make(map[string]interface{})

	baseViewModel := h.BaseViewModel.ViewModelForInlinePreviewAuthFlow(r, w)
	viewmodels.Embed(data, baseViewModel)

	screenViewModel := AuthflowCreatePasswordViewModel{
		AuthenticationStage:     string(authn.AuthenticationStagePrimary),
		PasswordManagerUsername: "",
		ForgotPasswordInputType: "",
		ForgotPasswordLoginID:   "",
	}

	passwordPolicyViewModel := viewmodels.NewPasswordPolicyViewModelFromAuthflow(
		declarative.NewPasswordPolicy(h.FeatureConfig.Authenticator, h.AuthenticatorConfig.Password.Policy),
		baseViewModel.RawError,
		&viewmodels.PasswordPolicyViewModelOptions{
			IsNew: true,
		},
	)

	passwordInputErrorViewModel := authflowv2viewmodels.NewPasswordInputErrorViewModel(baseViewModel.RawError)

	viewmodels.Embed(data, screenViewModel)
	viewmodels.Embed(data, passwordPolicyViewModel)
	viewmodels.Embed(data, passwordInputErrorViewModel)

	branchViewModel := h.InlinePreviewAuthflowBranchViewModeler.NewAuthflowBranchViewModelForInlinePreviewCreatePassword()
	viewmodels.Embed(data, branchViewModel)

	return data, nil
}

func (h *AuthflowV2CreatePasswordHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var handlers handlerwebapp.AuthflowControllerHandlers
	handlers.Get(func(s *webapp.Session, screen *webapp.AuthflowScreenWithFlowResponse) error {
		data, err := h.GetData(w, r, s, screen)
		if err != nil {
			return err
		}

		h.Renderer.RenderHTML(w, r, TemplateWebAuthflowCreatePasswordHTML, data)
		return nil
	})
	handlers.PostAction("", func(s *webapp.Session, screen *webapp.AuthflowScreenWithFlowResponse) error {
		err := AuthflowCreatePasswordSchema.Validator().ValidateValue(handlerwebapp.FormToJSON(r.Form))
		if err != nil {
			return err
		}

		index := *screen.Screen.TakenBranchIndex
		flowResponse := screen.BranchStateTokenFlowResponse

		screenData := flowResponse.Action.Data.(declarative.CreateAuthenticatorData)
		option := screenData.Options[index]

		newPlainPassword := r.Form.Get("x_password")
		confirmPassword := r.Form.Get("x_confirm_password")
		err = pwd.ConfirmPassword(newPlainPassword, confirmPassword)
		if err != nil {
			return err
		}

		input := map[string]interface{}{
			"authentication": option.Authentication,
			"new_password":   newPlainPassword,
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
		h.Renderer.RenderHTML(w, r, TemplateWebAuthflowCreatePasswordHTML, data)
		return nil
	})

	h.Controller.HandleStep(w, r, &handlers)
}
