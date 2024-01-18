package authflowv2

import (
	"net/http"

	handlerwebapp "github.com/authgear/authgear-server/pkg/auth/handler/webapp"
	"github.com/authgear/authgear-server/pkg/auth/handler/webapp/viewmodels"
	"github.com/authgear/authgear-server/pkg/auth/webapp"
	"github.com/authgear/authgear-server/pkg/lib/authenticationflow/declarative"
	"github.com/authgear/authgear-server/pkg/lib/authn"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator/password"
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
	Controller    *handlerwebapp.AuthflowController
	BaseViewModel *viewmodels.BaseViewModeler
	Renderer      handlerwebapp.Renderer
}

func (h *AuthflowV2CreatePasswordHandler) GetData(w http.ResponseWriter, r *http.Request, s *webapp.Session, screen *webapp.AuthflowScreenWithFlowResponse) (map[string]interface{}, error) {
	data := make(map[string]interface{})

	baseViewModel := h.BaseViewModel.ViewModelForAuthFlow(r, w)
	viewmodels.Embed(data, baseViewModel)

	screenViewModel := AuthflowCreatePasswordViewModel{}

	index := *screen.Screen.TakenBranchIndex
	flowResponse := screen.BranchStateTokenFlowResponse
	screenData := flowResponse.Action.Data.(declarative.IntentSignupFlowStepCreateAuthenticatorData)
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

	// put policy PasswordBelowGuessableLevel to front
	for p, x := range passwordPolicyViewModel.PasswordPolicies {
		if x.Name == password.PasswordBelowGuessableLevel {
			passwordPolicyViewModel.PasswordPolicies = append([]password.Policy{x}, append((passwordPolicyViewModel.PasswordPolicies)[:p], (passwordPolicyViewModel.PasswordPolicies)[p+1:]...)...)
			break
		}
	}

	viewmodels.Embed(data, screenViewModel)
	viewmodels.Embed(data, passwordPolicyViewModel)

	branchViewModel := viewmodels.NewAuthflowBranchViewModel(screen)
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
		data := flowResponse.Action.Data.(declarative.IntentSignupFlowStepCreateAuthenticatorData)
		option := data.Options[index]

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
	h.Controller.HandleStep(w, r, &handlers)
}
