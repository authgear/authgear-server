package authflowv2

import (
	"net/http"

	handlerwebapp "github.com/authgear/authgear-server/pkg/auth/handler/webapp"
	"github.com/authgear/authgear-server/pkg/auth/handler/webapp/viewmodels"
	"github.com/authgear/authgear-server/pkg/auth/webapp"
	"github.com/authgear/authgear-server/pkg/lib/authenticationflow/declarative"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	pwd "github.com/authgear/authgear-server/pkg/util/password"
	"github.com/authgear/authgear-server/pkg/util/template"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

var TemplateWebAuthflowChangePasswordHTML = template.RegisterHTML(
	"web/authflowv2/change_password.html",
	handlerwebapp.Components...,
)

var AuthflowChangePasswordSchema = validation.NewSimpleSchema(`
	{
		"type": "object",
		"properties": {
			"x_new_password": { "type": "string" },
			"x_confirm_password": { "type": "string" }
		},
		"required": ["x_new_password", "x_confirm_password"]
	}
`)

func ConfigureAuthflowv2ChangePasswordRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "POST", "GET").
		WithPathPattern(AuthflowV2RouteChangePassword)
}

type AuthflowV2ChangePasswordNavigator interface {
	NavigateChangePasswordSuccessPage(s *webapp.AuthflowScreen, r *http.Request, webSessionID string) (result *webapp.Result)
}

type AuthflowV2ChangePasswordHandler struct {
	Controller              *handlerwebapp.AuthflowController
	Navigator               AuthflowV2ChangePasswordNavigator
	BaseViewModel           *viewmodels.BaseViewModeler
	ChangePasswordViewModel *viewmodels.ChangePasswordViewModeler
	Renderer                handlerwebapp.Renderer
}

func (h *AuthflowV2ChangePasswordHandler) GetData(w http.ResponseWriter, r *http.Request, s *webapp.Session, screen *webapp.AuthflowScreenWithFlowResponse) (map[string]interface{}, error) {
	data := make(map[string]interface{})

	baseViewModel := h.BaseViewModel.ViewModelForAuthFlow(r, w)
	viewmodels.Embed(data, baseViewModel)

	screenData := screen.StateTokenFlowResponse.Action.Data.(declarative.ForceChangePasswordData)

	passwordPolicyViewModel := viewmodels.NewPasswordPolicyViewModelFromAuthflow(
		screenData.PasswordPolicy,
		baseViewModel.RawError,
		&viewmodels.PasswordPolicyViewModelOptions{
			IsNew: false,
		},
	)
	changePasswordViewModel := h.ChangePasswordViewModel.NewWithAuthflow(screenData.ForceChangeReason)

	viewmodels.Embed(data, passwordPolicyViewModel)
	viewmodels.Embed(data, changePasswordViewModel)

	return data, nil
}

func (h *AuthflowV2ChangePasswordHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var handlers handlerwebapp.AuthflowControllerHandlers
	handlers.Get(func(s *webapp.Session, screen *webapp.AuthflowScreenWithFlowResponse) error {
		data, err := h.GetData(w, r, s, screen)
		if err != nil {
			return err
		}

		h.Renderer.RenderHTML(w, r, TemplateWebAuthflowChangePasswordHTML, data)
		return nil
	})
	handlers.PostAction("", func(s *webapp.Session, screen *webapp.AuthflowScreenWithFlowResponse) error {
		err := AuthflowChangePasswordSchema.Validator().ValidateValue(handlerwebapp.FormToJSON(r.Form))
		if err != nil {
			return err
		}

		newPlainPassword := r.Form.Get("x_new_password")
		confirmPassword := r.Form.Get("x_confirm_password")
		err = pwd.ConfirmPassword(newPlainPassword, confirmPassword)
		if err != nil {
			return err
		}

		input := map[string]interface{}{
			"new_password": newPlainPassword,
		}

		result, err := h.Controller.AdvanceWithInput(r, s, screen, input, nil)
		if err != nil {
			return err
		}
		newScreen, err := h.Controller.DelayScreen(r, s, screen.Screen, result)
		if err != nil {
			return err
		}

		newResult := h.Navigator.NavigateChangePasswordSuccessPage(newScreen, r, s.ID)
		newResult.WriteResponse(w, r)
		return nil
	})
	h.Controller.HandleStep(w, r, &handlers)
}
