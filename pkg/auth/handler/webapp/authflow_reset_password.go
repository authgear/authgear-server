package webapp

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/auth/handler/webapp/viewmodels"
	"github.com/authgear/authgear-server/pkg/auth/webapp"
	"github.com/authgear/authgear-server/pkg/lib/authenticationflow/declarative"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	pwd "github.com/authgear/authgear-server/pkg/util/password"
	"github.com/authgear/authgear-server/pkg/util/template"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

var TemplateWebAuthflowResetPasswordHTML = template.RegisterHTML(
	"web/authflow_reset_password.html",
	components...,
)

var AuthflowResetPasswordSchema = validation.NewSimpleSchema(`
{
  "type": "object",
  "properties": {
    "x_password": { "type": "string" },
    "x_confirm_password": { "type": "string" }
  },
  "required": ["x_password", "x_confirm_password"]
}
`)

func ConfigureAuthflowResetPasswordRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "POST", "GET").
		WithPathPattern(webapp.AuthflowRouteResetPassword)
}

type AuthflowResetPasswordHandler struct {
	Controller    *AuthflowController
	BaseViewModel *viewmodels.BaseViewModeler
	Renderer      Renderer
}

func (h *AuthflowResetPasswordHandler) GetData(w http.ResponseWriter, r *http.Request, s *webapp.Session, screen *webapp.AuthflowScreenWithFlowResponse) (map[string]interface{}, error) {
	data := make(map[string]interface{})

	baseViewModel := h.BaseViewModel.ViewModelForAuthFlow(r, w)
	viewmodels.Embed(data, baseViewModel)

	screenData := screen.StateTokenFlowResponse.Action.Data.(declarative.IntentAccountRecoveryFlowStepResetPasswordData)

	passwordPolicyViewModel := viewmodels.NewPasswordPolicyViewModelFromAuthflow(
		screenData.PasswordPolicy,
		baseViewModel.RawError,
		&viewmodels.PasswordPolicyViewModelOptions{
			IsNew: false,
		},
	)
	viewmodels.Embed(data, passwordPolicyViewModel)

	return data, nil
}

func (h *AuthflowResetPasswordHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var handlers AuthflowControllerHandlers
	handlers.Get(func(s *webapp.Session, screen *webapp.AuthflowScreenWithFlowResponse) error {

		data, err := h.GetData(w, r, s, screen)
		if err != nil {
			return err
		}

		h.Renderer.RenderHTML(w, r, TemplateWebAuthflowResetPasswordHTML, data)
		return nil
	})
	handlers.PostAction("", func(s *webapp.Session, screen *webapp.AuthflowScreenWithFlowResponse) error {
		err := AuthflowResetPasswordSchema.Validator().ValidateValue(FormToJSON(r.Form))
		if err != nil {
			return err
		}

		newPassword := r.Form.Get("x_password")
		confirmPassword := r.Form.Get("x_confirm_password")
		err = pwd.ConfirmPassword(newPassword, confirmPassword)
		if err != nil {
			return err
		}

		result, err := h.Controller.AdvanceWithInput(r, s, screen, map[string]interface{}{
			"new_password": newPassword,
		})

		if err != nil {
			return err
		}

		result.WriteResponse(w, r)
		return nil
	})

	code := r.URL.Query().Get("code")
	h.Controller.HandleResumeOfFlow(w, r, webapp.SessionOptions{}, &handlers, map[string]interface{}{
		"account_recovery_code": code,
	})
}
